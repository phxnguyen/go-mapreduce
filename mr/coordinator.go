package mr

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type CoordinatorState int

const (
	Idle CoordinatorState = iota
	Map
	Reduce
	Done
)

type TaskState int

const (
	Wait TaskState = iota
	InProgress
	Completed
)

type taskInfo struct {
	task      Task
	state     TaskState
	startTime time.Time
}

type Coordinator struct {
	mu          sync.Mutex
	files       []string
	nReduce     int
	nMap        int
	mapDone     int
	reduceDone  int
	mapTasks    []taskInfo
	reduceTasks []taskInfo
	phase       CoordinatorState
}

// Your code here -- RPC handlers for the worker to call.

// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
func (this *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

// start a thread that listens for RPCs from worker.go
func (this *Coordinator) server() {
	rpc.Register(this)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (this *Coordinator) Done() bool {
	this.mu.Lock()

	defer this.mu.Unlock()

	return this.phase == Done
}

func (this *Coordinator) createMapTask() {
	this.mapTasks = make([]taskInfo, len(this.files))

	for idx, filename := range this.files {
		this.mapTasks[idx] = taskInfo{
			task: Task{
				Type:      MapTask,
				TaskId:    idx,
				Filename:  filename,
				NumReduce: this.nReduce,
			},
			state: Wait,
		}
	}
}

func (this *Coordinator) createReduceTask() {
	this.reduceTasks = make([]taskInfo, this.nReduce)
	for idx := range this.reduceTasks {
		mapfiles := make([]string, 0, this.nMap)
		for j := 0; j < this.nMap; j++ {
			mapfiles = append(mapfiles, fmt.Sprintf("mr-%d-%d", j, idx))
		}
		this.reduceTasks[idx] = taskInfo{
			task: Task{
				Type: ReduceTask, TaskId: idx, NumReduce: this.nReduce, MapFiles: mapfiles,
			},
			state: Wait,
		}
	}
}

func (this *Coordinator) AssignTask(args *RequestTaskArgs, reply *RequestTaskReply) error {
	this.mu.Lock()
	defer this.mu.Unlock()

	if this.phase == Map {
		if task, ok := this.getNextTask(this.mapTasks); ok {
			reply.Task = task
			return nil
		}

		if this.mapDone == this.nMap {
			this.phase = Reduce
		} else {
			reply.Task = Task{Type: WaitTask}
			return nil
		}
	}

	if this.phase == Reduce {
		if task, ok := this.getNextTask(this.reduceTasks); ok {
			reply.Task = task
			return nil
		}
		if this.reduceDone == this.nReduce {
			this.phase = Done
		} else {
			reply.Task = Task{Type: WaitTask}
			return nil
		}
	}

	reply.Task = Task{Type: ExitTask}
	return nil
}

func (this *Coordinator) getNextTask(tasks []taskInfo) (Task, bool) {
	now := time.Now()

	// ưu tiên task idle
	for i := range tasks {
		if tasks[i].state == Wait {
			tasks[i].state = InProgress
			tasks[i].startTime = now
			return tasks[i].task, true
		}
	}

	// nếu không có task idle, thì lấy task in progress của mấy con stranggle worker
	for i := range tasks {
		if tasks[i].state == InProgress && now.Sub(tasks[i].startTime) > 10*time.Second {
			tasks[i].startTime = now
			return tasks[i].task, true
		}
	}

	return Task{}, false
}

func (this *Coordinator) CompleteTask(args *CompleteTaskArgs, reply *CompleteTaskReply) error {
	this.mu.Lock()
	defer this.mu.Unlock()
	switch args.TaskType {
	case MapTask:
		if args.TaskId >= 0 && args.TaskId < this.nMap &&
			this.mapTasks[args.TaskId].state != Completed {
			this.mapTasks[args.TaskId].state = Completed
			this.mapDone += 1
		}
	case ReduceTask:
		if args.TaskId >= 0 && args.TaskId < this.nReduce &&
			this.reduceTasks[args.TaskId].state != Completed {
			this.reduceTasks[args.TaskId].state = Completed
			this.reduceDone += 1
		}
	}
	return nil
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{
		nReduce: nReduce,
		nMap:    len(files),
		phase:   Map,
	}

	c.createMapTask()
	c.createReduceTask()
	c.server()
	return &c
}
