package rotatelogs_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	rotatelogs "github.com/iproj/file-rotatelogs"
	"github.com/stretchr/testify/assert"
)

func ExampleForceNewFile() {
	logDir, err := ioutil.TempDir("", "rotatelogs_test")
	if err != nil {
		fmt.Println("could not create log directory ", err)

		return
	}
	logPath := fmt.Sprintf("%s/test.log", logDir)

	for i := 0; i < 2; i++ {
		writer, err := rotatelogs.New(logPath,
			rotatelogs.ForceNewFile(),
		)
		if err != nil {
			fmt.Println("Could not open log file ", err)

			return
		}

		n, err := writer.Write([]byte("test"))
		if err != nil || n != 4 {
			fmt.Println("Write failed ", err, " number written ", n)

			return
		}
		err = writer.Close()
		if err != nil {
			fmt.Println("Close failed ", err)

			return
		}
	}

	files, err := ioutil.ReadDir(logDir)
	if err != nil {
		fmt.Println("ReadDir failed ", err)

		return
	}
	for _, file := range files {
		fmt.Println(file.Name(), file.Size())
	}

	err = os.RemoveAll(logDir)
	if err != nil {
		fmt.Println("RemoveAll failed ", err)

		return
	}
	// OUTPUT:
	// test.log 4
	// test.log.1 4
}

func TestTooMuchLog(t *testing.T) {

	var (
		testDir       = "test_much_log"
		testLogPath   = filepath.Join(testDir, "access_log")
		rotationCount = 3
		N             = 12 // N > 10
	)
	err := os.Mkdir(testDir, 0777)
	assert.Nil(t, err)
	defer os.RemoveAll(testDir)
	assert.Nil(t, err)

	rl, err := rotatelogs.New(
		testLogPath+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(testLogPath),
		rotatelogs.WithRotationCount(uint(rotationCount)),
		rotatelogs.WithRotationSize(12), // Log contentSize > 12
	)
	assert.Nil(t, err)

	log.SetOutput(rl)
	for i := 0; i < N; i++ {
		log.Printf("Test content %d\n", i)
	}
	files, _ := ioutil.ReadDir(testDir)
	assert.Equal(t, rotationCount+1, len(files))

	bytez, err := ioutil.ReadFile(testLogPath)
	assert.Nil(t, err)
	assert.Equal(t, strconv.Itoa(N-1), string(bytez[len(bytez)-3:len(bytez)-1]))
}

func TestSizePriorityOverTime(t *testing.T) {

	var (
		testDir       = "test_size_priority_over_time"
		testLogPath   = filepath.Join(testDir, "access_log")
		rotationCount = 2
		N             = 12
	)
	err := os.Mkdir(testDir, 0777)
	assert.Nil(t, err)
	defer os.RemoveAll(testDir)
	assert.Nil(t, err)

	rl, err := rotatelogs.New(
		testLogPath+".%Y%m%d%H%M%S", // Accurate to seconds
		rotatelogs.WithRotationCount(uint(rotationCount)),
		rotatelogs.WithRotationSize(12000),
	)
	assert.Nil(t, err)

	log.SetOutput(rl)
	for i := 0; i < N; i++ {
		log.Printf("Test content %d\n", i)
		// N * sleepTime > 1s
		time.Sleep(120 * time.Millisecond)
	}
	files, _ := ioutil.ReadDir(testDir)
	assert.Equal(t, 1, len(files)) // N * contentSize < rotationSize

	rl, err = rotatelogs.New(
		testLogPath+".%Y%m%d%H%M%S", // Accurate to seconds
		rotatelogs.WithRotationCount(uint(rotationCount)),
	)
	assert.Nil(t, err)

	log.SetOutput(rl)
	for i := 0; i < N; i++ {
		log.Printf("Test content %d\n", i)
		// N * sleepTime > 1s
		time.Sleep(120 * time.Millisecond)
	}
	files, _ = ioutil.ReadDir(testDir)
	assert.Equal(t, rotationCount, len(files))
}
