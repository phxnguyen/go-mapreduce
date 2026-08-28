1. Vì sao không lưu data vào RAM cho nhanh?

- Mỗi máy có 2-4GB RAM (paper mục 3.1), nhưng một job MapReduce xử lý hàng terabyte data, sinh ra intermediate data cỡ 758 TB (theo Table 1, số liệu tháng 8/2004). Dù RAM có rẻ đến đâu, dữ liệu trung gian của toàn bộ cluster vẫn không thể nào nhét vừa vào memory tổng.
- Map task ghi kết quả ra local disk thay vì giữ trong RAM vì reduce worker có thể đọc dữ liệu ở thời điểm thích hợp — không phải đọc ngay lập tức. Nếu map task xong mà worker bị kill/crash trước khi mọi reduce worker kịp đọc xong, dữ liệu trong RAM sẽ mất sạch.
Paper nói rõ ở mục 3.3: "Completed map tasks are re-executed on a failure because their output is stored on the local disk(s) of the failed machine and is therefore inaccessible." — tức là ngay cả khi ghi ra disk rồi, nếu máy đó chết thì vẫn phải chạy lại map task, vì disk đó không còn truy cập được (khác với GFS — hệ thống phân tán có replication).
Nếu để trong RAM, xác suất mất dữ liệu khi máy chết còn cao hơn, và không có cách nào "khôi phục" ngoài chạy lại từ đầu.

Nói cách khác: ghi xuống local disk là checkpoint tạm thời