package agent

import (
	"os/exec"
	"syscall"
	"os"
)

func handoffToK3s(flags []string){
	k3sPath, err := exec.LookPath("k3s")
	if err != nil {
		panic(err)
	}

	args := append([]string{"agent"}, flags...)

	env := os.Environ()
	
	err = syscall.Exec(k3sPath, args, env)

	if err != nil {
		panic(err)
	}


	 
}
