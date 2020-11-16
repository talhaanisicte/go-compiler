package main

// #include <stdlib.h>
import "C"
import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"unsafe"
)

var mtx sync.Mutex

//export GetGasForData
func GetGasForData(ptr unsafe.Pointer, len C.int) uint64 {
	mtx.Lock()
	defer mtx.Unlock()
	return getGasForData(C.GoBytes(ptr, len))
}

//export Run
func Run(ptr unsafe.Pointer, len C.int) unsafe.Pointer {
	mtx.Lock()
	defer mtx.Unlock()
	rarr := run(C.GoBytes(ptr, len))
	cArr := C.CBytes(rarr)
	return cArr
}
func main() {}

/*///////////////////////////////////////////////////////////////////////////////
WARNING: DON'T MODIFY UPPER PART. QA TESTER WILL GENERATE AN ERROR AFTER SUBMSSTION
ONLY IMPORT SECTION CAN BE MODIFIED.
/////////////////////////////////////////////////////////////////////////////////*/

// getGasForData - Returns back gas required to execute the contract
func getGasForData([]byte) uint64 {
	// calculate gas here
	return uint64(5000000)
}

// run - Runs the contract, It recieve data as parsed byte and returns back a parsed byte array
func run(arr []byte) []byte {
	// Example of returning time in byte array
	repoLink := string(arr)
	fmt.Println(repoLink)
	err := clone(repoLink)
	if err != nil {
		return getBytes([]byte{}, errors.New("Clone err"))
	}
	err = compile()
	if err != nil {
		return getBytes([]byte{}, errors.New("Compile err"))
	}
	hash, err := publishToIPFS()
	if err != nil {
		return getBytes([]byte{}, errors.New("Publish err"))
	}
	return getBytes(hash, err)
}

var goRoot string = os.Getenv("HOME") + "/go/src/"
var repoRoot string = ""

func getBytes(msg []byte, err error) []byte {
	msgLenBytes := make([]byte, 4)
	if err != nil {
		binary.BigEndian.PutUint32(msgLenBytes, uint32(len(err.Error())))
		return append(append([]byte{0, 253, 253}, msgLenBytes...), []byte(err.Error())...)
	}
	binary.BigEndian.PutUint32(msgLenBytes, uint32(len(msg)))
	return append(append([]byte{1, 253, 253}, msgLenBytes...), msg...)
}

func check(err error) bool {
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func clone(repoLink string) error {
	if !strings.HasSuffix(repoLink, ".git") {
		repoLink = repoLink + ".git"
	}
	repoPath := strings.Split(strings.Split(repoLink, "https://")[1], ".git")[0]
	repoRoot = goRoot + repoPath
	_, err := os.Stat(repoRoot)
	if !os.IsNotExist(err) {
		err := os.RemoveAll(repoRoot)
		if !check(err) {
			return err
		}
	}
	cmd := exec.Command("git", "clone", repoLink, repoRoot)
	_, err = cmd.CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}

func compile() error {
	cmd := exec.Command("go", "get", "./...")
	cmd.Dir = repoRoot
	_, err := cmd.CombinedOutput()
	if !check(err) {
		return err
	}
	cmd = exec.Command("go", "build", "-o", "goLib.so", "-buildmode=c-shared")
	cmd.Dir = repoRoot
	_, err = cmd.CombinedOutput()
	if !check(err) {
		return err
	}
	return nil
}

func publishToIPFS() ([]byte, error) {
	cmd := exec.Command("ipfs", "add", "-Q", "goLib.so")
	cmd.Dir = repoRoot
	hash, err := cmd.CombinedOutput()
	strHash := strings.TrimSuffix(string(hash), "\n")
	if err != nil {
		return []byte{}, err
	}
	fmt.Println(strHash)
	return []byte(strHash), nil
}
