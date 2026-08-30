# go-mapreduce

An implementation of Lab 1 (MapReduce) from MIT's [6.824/6.5840 Distributed Systems](http://nil.csail.mit.edu/6.824/2021/labs/lab-mr.html) course (Spring 2021).

Starter code © MIT 6.824 staff; provided here for personal coursework/learning purposes only.

---

MapReduce tổng quan

MapReduce quy trình thực được nêu rõ ở phần 3.1 các bước như sau 

> 1. The MapReduce library in the user program first
     splits the input files into M pieces of typically 16
     megabytes to 64 megabytes (MB) per piece (controllable by the user via an optional parameter). It
     then starts up many copies of the program on a cluster of machines.

MapReduce lib sẽ chia input file thành *M* phần, mỗi phần có kích thước 16-64 MB và có thể thay đổi tùy ý. Các phần chia nhỏ này sẽ chạy các cụm worker 

> 2. One of the copies of the program is special – the
master. The rest are workers that are assigned work
by the master. There are M map tasks and R reduce
tasks to assign. The master picks idle workers and
assigns each one a map task or a reduce task.

Trong cả đám worker đó sẽ có một master. Tất cả workers còn lại được phân bổ công việc bởi master. 
Có $M$ map tasks và $R$ reduce tasks được assign. Master sẽ chọn các workers nào rảnh(idle) để assign một map task hoặc một reduce task

> 3. A worker who is assigned a map task reads the contents of the corresponding input split. 
> It parses key/value pairs out of the input data and passes each pair to the user-defined Map function. 
> The intermediate key/value pairs produced by the Map function are buffered in memory.

Một worker được assign một map task sẽ đọc contents từ phần input đã được split 

Worker parses cặp key/val khỏi input data và áp dụng logic Map function mà user định nghĩa 

Sau khi áp logic Map lên từng cặp k/v, kết quả sẽ được buffer in memory - gom lại trên memory 


> 4. Periodically, the buffered pairs are written to local
     disk, partitioned into R regions by the partitioning
     function. The locations of these buffered pairs on
     the local disk are passed back to the master, who
     is responsible for forwarding these locations to the
     reduce workers.

Định kỳ, kết quả từ Map function (intemediate) hay buffer in memory mới được wirte (flush) xuống tại local disk, chia thành $R$ partition theo partitioning function

> 5. When a reduce worker is notified by the master
     about these locations, it uses remote procedure calls
     to read the buffered data from the local disks of the
     map workers. When a reduce worker has read all intermediate data, it sorts it by the intermediate keys
     so that all occurrences of the same key are grouped
     together. The sorting is needed because typically
     many different keys map to the same reduce task. If
     the amount of intermediate data is too large to fit in
     memory, an external sort is used.

Khi một reduce worker nhận thông báo từ master về nơi chứa intermediate data, reduce worker sẽ dùng RPC để đọc intermediate data từ local disks của chính map worker. 

Khi reduce worker remote read xong intermediate data từ nhiều local disk của các map worker, data lúc này lộn xộn. Reduce worker **sort** toàn bô theo key. Mục đích để tất cả value có key nằm cạnh nhau, thay vì nằm rải rác.

Trong trường hợp intermediate quá lớn không vừa memory thì dùng external sort (ví dụ merge sort trên disk). Đây là lý do ở 4.2 "Ordering Guarantees", nói là trong mỗi partition (kết quả của reduce task) đảm bảo order, vì đã được sort ở bước này

> 6. The reduce worker iterates over the sorted intermediate data and for each unique intermediate key encountered, it passes the key and the corresponding set of intermediate values to the user’s Reduce function. The output of the Reduce function is appended
to a final output file for this reduce partition.


Reduce worker interate over dữ liệu đã sort, cứ mỗi lần gặp một key mới thì gom hết các value cùng key đó lại thành một danh sách, gọi hàm Reduce(key, listValue) để xử lý, rồi ghi kết quả trả về vào file output. (gần giống bài Group Anagrams )

> 7. When all map tasks and reduce tasks have been
     completed, the master wakes up the user program.
     At this point, the MapReduce call in the user program returns back to the user code.

Khi map task và reduce task đã xong, master sẽ gọi user program. Lúc này MapReduce return kết quả 