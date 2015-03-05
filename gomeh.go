package main

import (
	"os"
	"io/ioutil"
	"fmt"
	"log"
	"runtime"
	"strings"
	"errors"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"

	gj "github.com/kpawlik/geojson"
	"encoding/json"

	"github.com/StefanSchroeder/Golang-Ellipsoid/ellipsoid"
)

const (
	WindowWidth = 1000
	WindowHeight = 1000
)

func loadPoints(filename string) []float64 {
	var feature gj.Feature
	geo1 := ellipsoid.Init("WGS84", ellipsoid.Degrees, ellipsoid.Meter, ellipsoid.LongitudeIsSymmetric, ellipsoid.BearingIsSymmetric)

	file, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(file, &feature)
	if err != nil {
		log.Fatal(err)
	}

	g, err := feature.GetGeometry()
	if err != nil {
		log.Fatal(err)
	}

	mp, ok := g.(*gj.MultiPoint)
	if !ok {
		log.Fatalln("Failed to extract point coordinates from geojson")
	}

	fmt.Println("Loaded point coordinates:", len(mp.Coordinates))

	pointSlice := make([]float64, len(mp.Coordinates)*3)
	i := 0
	for  _, p := range mp.Coordinates {
		x, y, z := geo1.ToECEF(float64(p[1]), float64(p[0]), 0)
		pointSlice[i + 0] = x / geo1.Ellipse.Equatorial
		pointSlice[i + 1] = y / geo1.Ellipse.Equatorial
		pointSlice[i + 2] = z / geo1.Ellipse.Equatorial
		i += 3
	}

	return pointSlice
}

func main() {
	var x float32
	var y float32
	var span float32 = 2

	if len(os.Args) != 2 {
		fmt.Printf("usage: %s <filename>\n", os.Args[0])
		os.Exit(1)
	}

	pointSlice := loadPoints(os.Args[1])

	runtime.LockOSThread()
	if err := glfw.Init(); err != nil {
		log.Fatalln("Failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	window, err := glfw.CreateWindow(WindowWidth, WindowHeight, "Cube", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	window.SetKeyCallback(func (w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if key == glfw.KeyA && action == glfw.Press {
			x -= span/10
		}

		if key == glfw.KeyD && action == glfw.Press {
			x += span/10
		}

		if key == glfw.KeyS && action == glfw.Press {
			y -= span/10
		}

		if key == glfw.KeyW && action == glfw.Press {
			y += span/10
		}

		if key == glfw.KeyR && action == glfw.Press {
			span /= 2
		}

		if key == glfw.KeyF && action == glfw.Press {
			span *= 2
		}
	})

	if err := gl.Init(); err != nil {
		log.Fatal("Failed to initialize opengl")
	}

	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL version", version)

	program, err := newProgram(vertexShader, fragmentShader)
	if err != nil {
		log.Fatal(err)
	}
	gl.UseProgram(program)

	projection := mgl32.Ortho(-span/2, span/2, -span/2, span/2, -span/2, span/2)
	projectionUniform := gl.GetUniformLocation(program, gl.Str("projection\x00"))
	gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])

	camera := mgl32.Ident4()
	cameraUniform := gl.GetUniformLocation(program, gl.Str("camera\x00"))
	gl.UniformMatrix4fv(cameraUniform, 1, false, &camera[0])

	model := mgl32.Ident4()
	modelUniform := gl.GetUniformLocation(program, gl.Str("model\x00"))
	gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(pointSlice) * 8, gl.Ptr(pointSlice), gl.STATIC_DRAW)

	vertAttrib := uint32(gl.GetAttribLocation(program, gl.Str("vert\x00")))
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointer(vertAttrib, 3, gl.DOUBLE, false, 0, gl.PtrOffset(0))

	gl.ClearColor(0.0, 0.0, 0.0, 0.0)
	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		projection := mgl32.Ortho(-span/2, span/2, -span/2, span/2, -100, 100)
		gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])

		camera = mgl32.Translate3D(-x, -y, -5)
		gl.UniformMatrix4fv(cameraUniform, 1, false, &camera[0])

		gl.BindVertexArray(vao)
		gl.DrawArrays(gl.POINTS, 0, int32(len(pointSlice)/3))

		window.SwapBuffers()
		glfw.PollEvents()
	}

}
func newProgram(vertexShaderSource, fragmentShaderSource string) (uint32, error) {
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	program := gl.CreateProgram()

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, errors.New(fmt.Sprintf("failed to link program: %v", log))
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csource := gl.Str(source)
	gl.ShaderSource(shader, 1, &csource, nil)
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

var vertexShader string = `
#version 330
uniform mat4 projection;
uniform mat4 camera;
uniform mat4 model;
in vec3 vert;
void main() {
    gl_Position = projection * camera * model * vec4(vert, 1);
}
` + "\x00"

var fragmentShader = `
#version 330
out vec4 outputColor;
void main() {
    outputColor = vec4(1.0, 1.0, 1.0, 1.0);
}
` + "\x00"
