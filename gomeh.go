package main

import (
	"os"
	"io/ioutil"
	"fmt"
	"log"
	"encoding/json"
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
	if len(os.Args) != 2 {
		fmt.Printf("usage: %s <filename>\n", os.Args[0])
		os.Exit(1)
	}

	d := loadJson(os.Args[1])

	fmt.Println(d.Type.Point)
}
