package main

import (
	"os"
	"io/ioutil"
	"fmt"
	"log"
	"encoding/json"
	"runtime"
	"strings"
	"errors"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	WindowWidth = 1000
	WindowHeight = 1000
)

type pointsAndBoxes struct {
	Type struct {
		Box map[string]interface{}
		Point map[string]struct {
			Lambda float64 `json:",string"`
			Phi float64 `json:",string"`
		}
	}
}

func loadJson(filename string) (ret *pointsAndBoxes) {
	ret = &pointsAndBoxes{}
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	json.Unmarshal(file, ret)
	return ret
}

func main() {
	var x float32
	var y float32

	if len(os.Args) != 2 {
		fmt.Printf("usage: %s <filename>\n", os.Args[0])
		os.Exit(1)
	}
	d := loadJson(os.Args[1])

	fmt.Println("Number of points:", len(d.Type.Point))
	pointSlice := make([]float64, len(d.Type.Point)*2)
	i := 0
	for  _, p := range d.Type.Point {
		pointSlice[i] = p.Phi
		pointSlice[i + 1] = p.Lambda
		i += 2
	}

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
			x -= 1
		}

		if key == glfw.KeyD && action == glfw.Press {
			x += 1
		}

		if key == glfw.KeyS && action == glfw.Press {
			y -= 1
		}

		if key == glfw.KeyW && action == glfw.Press {
			y += 1
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

//	projection := mgl32.Ortho2D(-90, 90, -90, 90)
	projection := mgl32.Ortho2D(10, 25, 55, 70)
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
	gl.VertexAttribPointer(vertAttrib, 2, gl.DOUBLE, false, 0, gl.PtrOffset(0))

	gl.ClearColor(0.0, 0.0, 0.0, 0.0)
	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		camera := mgl32.Translate3D(-x, -y, 0)
		gl.UniformMatrix4fv(cameraUniform, 1, false, &camera[0])

		gl.BindVertexArray(vao)
		gl.DrawArrays(gl.POINTS, 0, int32(len(pointSlice)/2))

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
in vec2 vert;
void main() {
    gl_Position = projection * camera * model * vec4(vert, 0, 1);
}
` + "\x00"

var fragmentShader = `
#version 330
out vec4 outputColor;
void main() {
    outputColor = vec4(1.0, 1.0, 1.0, 1.0);
}
` + "\x00"
