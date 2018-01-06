package main

import (
	"fmt"
	"os/exec"
)

func main() {
	python, err := exec.LookPath("python")
	if err != nil {
		fmt.Println("No python installed")
		return
	}

	fmt.Println("Python installed at ", python)
	target := "--content /Users/caoyuanle/Documents/neural-style/examples/1-content.jpg"
	style := "--styles /Users/caoyuanle/Documents/neural-style/examples/1-style.jpg"
	output := "--output /Users/caoyuanle/Documents/neural-style/examples/demo.jpg"
	iterationTimes := "--iterations 1"
	pyfiles := "/Users/caoyuanle/Documents/neural-style-server/neural-style/neural_style.py"
	cmd := exec.Command(python, pyfiles, target, style, output, iterationTimes)
	err = cmd.Start()
	if err != nil {
		fmt.Println("Bad command: ", err)
		return
	}

	err = cmd.Wait()
	//data, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error : ", err)
	} else {
		fmt.Println("Good Style Transfer")
	}
}
