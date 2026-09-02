package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"net/rpc"
	"os"
	"time"
)

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

type ByKey []KeyValue

func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

// main/mrworker.go calls this function.
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	// Your worker implementation here.

	// uncomment to send the Example RPC to the coordinator.
	// CallExample()

	for {
		task := getTask()
		switch task.Type {
		case MapTask:
			doMap(task.TaskId, task.Filename, task.NumReduce, mapf)
		case ReduceTask:
			doReduce(task.TaskId, task.MapFiles, reducef)
		case WaitTask:
			time.Sleep(6 * time.Second) // find out using sync.cond
		case ExitTask:
			return
		}
	}
}

func doMap(taskId int, filename string, numReduce int, mapf func(string, string) []KeyValue) {
	content, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("cannot read %v", filename)
	}

	intermediates := mapf(filename, string(content))

	buckets := make([][]KeyValue, numReduce)
	for _, intermediate := range intermediates {
		partition := ihash(intermediate.Key) % numReduce
		buckets[partition] = append(buckets[partition], intermediate)
	}
	for i, bucket := range buckets {
		oname := fmt.Sprintf("mr-%d-%d", taskId, i)
		tmp, err := os.CreateTemp("", "mr-tmp-*")
		if err != nil {
			log.Fatalf("cannot create temp file: %v", err)
		}
		enc := json.NewEncoder(tmp)
		for _, kv := range bucket {
			if err := enc.Encode(&kv); err != nil {
				log.Fatalf("cannot encode %v", kv)
			}
		}
		tmp.Close()
		os.Rename(tmp.Name(), oname)
	}

	CompleteTask(taskId, MapTask)
}

func doReduce(taskId int, mapFiles []string, reducef func(string, []string) string) {}

func getTask() Task {
	args := RequestTaskArgs{}
	reply := RequestTaskReply{}
	ok := call("Coordinator.AssignTask", &args, &reply)
	if !ok {
		return Task{Type: ExitTask}
	}
	return reply.Task
}

func CompleteTask(taskId int, taskType TaskType) {
	args := CompleteTaskArgs{
		TaskId:   taskId,
		TaskType: taskType,
	}
	reply := CompleteTaskReply{}
	ok := call("Coordinator.CompleteTask", &args, &reply)
	if !ok {
		log.Fatalf("failed to complete task: %v", taskId)
	}
}

// example function to show how to make an RPC call to the coordinator.
//
// the RPC argument and reply types are defined in rpc.go.
func CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	call("Coordinator.Example", &args, &reply)

	// reply.Y should be 100.
	fmt.Printf("reply.Y %v\n", reply.Y)
}

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		return false
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
