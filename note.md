# MapReduce Notes

## 1. Vì sao không lưu data đã được map vào RAM cho nhanh?

Mỗi máy có 2-4GB RAM nhưng một job MapReduce xử lý cả terabyte data, sinh ra intermediate data cỡ 758 TB nên không thể nào nhét vừa vào memory.

Map task ghi kết quả ra local disk thay vì giữ trong RAM vì reduce worker đọc file này qua RPC và ở một thời điểm nào đó (không biết khi nào nhưng không phải ngay liền).

Paper nói rõ ở mục 3.3: *"Completed map tasks are re-executed on a failure because their output is stored on the local disk(s) of the failed machine and is therefore inaccessible."* — tức là ngay cả khi ghi ra disk rồi, nếu máy đó chết thì vẫn phải chạy lại map task, vì disk đó không còn truy cập được (khác với GFS — hệ thống phân tán có replication).

Để trong memory thì nó không persistence.

## 2. Map function có thể chạy lại (re-execute) không?

**Context:** mỗi `worker` làm một task trong một thời điểm, xong task thì gửi completion message và ask `coordinator` (master) for next task.

Song với đó, `coordinator` ping định kỳ tới các worker kiểm tra còn sống không.

Nếu network partition xảy ra, dù task worker đã xong nhưng message cả đi và về đều không có, master đánh dấu worker đó là failed, và reset các task nó đang giữ (dù kể cả map task đã complete) về idle để giao cho worker khác chạy lại.

MapReduce mượn ý tưởng từ functional programming, function luôn predict được result (deterministic/functional), same input thì same result. Điều này dẫn đến fault tolerance, khỏi checkpoint, khỏi rollback mệt người, tưởng lỗi mà lại là tính năng. Thế mà lại hay 😎

## 3. Reduce có thể chạy lại (re-execute) không?

**Context:** same as **2.**

Reduce vẫn có mechanism rerun y hệt map function, there is no difference 😊

## 4. Map worker có thể fail, Reduce worker có thể fail, thế Coordinator có thể fail không?

Không, "không" ở đây có nghĩa là nó không nên fail, không được thiết kế để fail, và không mong là nó fail. Nhưng nó vẫn có thể fail 😭

Khi coordinator fail, nó không có mechanism re-execute, client (user gọi lib MapReduce) phải chạy lại, hoặc có thể nói re-execute bằng cơm.

## 5. Worker chạy chậm thì tính sao?

Worker chạy chậm (straggler) thì không cho là failure nên không re-execution. Coordinator giao cùng một task đó cho một worker khác chạy song song, worker nào xong trước thì lấy kết quả của worker đó.
