package job

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/elastic/beats/libbeat/common"

	"techtoolkit.ickey.cn/cockimport/worker"
)

// Job represents the job to be run
type Job struct {
	Name       string
	Wg         *sync.WaitGroup
	Rnd        *rand.Rand
	Event      *common.MapStr
	JobQueue   chan worker.Job
	RandomFail bool
	Retries    int
	Attempts   int
}

func (j Job) String() string {
	return j.Name
}

func (j Job) Process() error {
	// fmt.Printf("Processing job %s  event: %v\n", j.String(), j.Event)
	fmt.Printf("Processing job %s  event: %v\n", j.String(), *j.Event)
	if !j.RandomFail || j.Rnd.Intn(2) == 0 {
		return nil
	}
	return fmt.Errorf("Some kind of error")
}

func (j Job) Result(err error) {
	if err != nil {
		fmt.Printf("Processing job %s failed: %s\n", j.String(), err)
		if j.Attempts < j.Retries {
			j.Attempts++
			fmt.Printf("Retrying job %s [Retries:%d]\n", j.String(), j.Attempts)
			j.JobQueue <- j
			return
		} else {
			fmt.Printf("Giving up job %s [Retries:%d]\n", j.String(), j.Attempts)
		}
	} else {
		fmt.Printf("Processed job %s successfully\n", j.String())
	}
	j.Wg.Done()
}
