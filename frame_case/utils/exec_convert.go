package utils

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

/*
func exe_cmd(cmd string, wg *sync.WaitGroup) {
	fmt.Println("command is ", cmd)
	// splitting head => g++ parts => rest of the command
	parts := strings.Split(cmd, "!")
	fmt.Println(parts)
	head := parts[0]
	args := parts[1:len(parts)]
	fmt.Println(args)
	cmd_exec := exec.Command(head, args...)
	//	Sanity check -- capture stdout and stderr:
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd_exec.Stdout = &out
	cmd_exec.Stderr = &stderr

	//	Run the command
	cmd_exec.Run()

	//	Output our results
	fmt.Printf("Result: %v / %v \n", out.String(), stderr.String())
	wg.Done() // Need to signal to waitgroup that this goroutine is done
}

*/

func exe_cmd(cmd string, wg *sync.WaitGroup) {
	fmt.Println("command is ", cmd)
	// splitting head => g++ parts => rest of the command
	parts := strings.Split(cmd, "?")
	for _, part := range parts {
		fmt.Println(part)
	}
	head := parts[0]
	args := parts[1:len(parts)]
	cmd_exec := exec.Command(head, args...)
	stdoutStderr, err := cmd_exec.CombinedOutput()
	if err != nil {
		// TODO: handle error more gracefully
		fmt.Println(err)
	}
	// do something with output
	fmt.Printf("%s\n", stdoutStderr)

	wg.Done() // Need to signal to waitgroup that this goroutine is done
}

/*
func main() {

	commands := []string{"convert?./img/Escam25_2019-03-17_21-15-31.jpeg?-gravity?Southeast?-draw?text 20,8 'one'?./img/Escam25.jpeg"} //"convert ./img/Escam25_2019-03-17_21-15-31.jpeg -pointsize 20 -draw "gravity southeast fill yellow text 20,8 'test_camera'" ./img/Escam_outs.jpeg
	//"} //"convert !./img/Escam25_2019-03-17_21-15-31.jpeg  !-pointsize 20  !-gravity southeast !-fill yellow    !./img/Escam25_2019-03-17_21-15-31.jpeg"}

	wg := new(sync.WaitGroup)
	for _, command := range commands {
		wg.Add(1)
		go exe_cmd(command, wg)

	}
	wg.Wait()

}
*/
