# Chronos Frontend (FE) — Codebase Intelligence Platform

Giao diện web hoàn chỉnh với giao diện **Dark Mode** cao cấp, tích hợp đầy đủ cả 5 module nghiệp vụ của hệ sinh thái Chronos.

---

## 🌟 5 Chức Năng Chính:

1. **Dashboard (Home Page):**
   - Thống kê tổng quan: Total Files, Functions, Vectors (Postgres), Graph Nodes (Neo4j), NATS Jobs.
   - Thẻ dự án (Repository Cards) hiển thị trạng thái `Indexed`.
   - Nút **"Add Repository"** mở modal kích hoạt luồng Ingestion (`/api/v1/ingest-dir`) với thanh tiến trình NATS.

2. **AI Chat (RAG Dual-Panel - F7):**
   - **Khung bên trái (Chat):** Stream câu trả lời từng chữ qua SSE (`/api/v1/ask-stream`), có nút Stop Stream.
   - **Chips trích dẫn:** Dưới mỗi câu trả lời có các chip `[1] ask.go:37`, `[2] git_service.go:3` có thể click.
   - **Khung bên phải (Code Viewer):** Khi bấm vào chip trích dẫn, code viewer lập tức mở file và highlight sáng rực các dòng mã nguồn tương ứng.

3. **Semantic Search (F6):**
   - Thanh tìm kiếm câu hỏi tự nhiên thuần túy bằng vector (`/api/v1/search`).
   - Danh sách code snippets kèm thanh hiển thị độ tương đồng Cosine Similarity (vd: `96% Match`).

4. **Knowledge Graph Explorer (Neo4j - F4 & F8):**
   - Canvas đồ thị tương tác mượt mà vẽ bằng **Cytoscape.js**.
   - Node Hàm (Màu xanh dương) và Node File (Màu cam) liên kết với nhau bằng các mũi tên `CALLS` và `BELONGS_TO`.
   - **Impact Analysis Tool (F8):** Nhập tên hàm (vd: `AskCmd`) bấm **"Blast Radius"** (`/api/v1/impact`). Node mục tiêu sẽ đỏ rực, các hàm bị ảnh hưởng (Callers) sẽ sáng vàng, các node không liên quan tự động mờ đi.
   - **Node Inspector:** Click vào bất kỳ node nào trên đồ thị để xem chi tiết vị trí dòng, file và kích hoạt nhanh Impact Analysis.

5. **System Health & Logs (Admin Panel - F9):**
   - Giám sát trạng thái 6 dịch vụ: Go Core, Python AI, Rust Parser, Postgres (pgvector), Neo4j, NATS.
   - Terminal mô phỏng logs realtime từ hàng đợi NATS.

---

## 🚀 Cách mở giao diện:

### Cách 1: Click đúp mở trực tiếp (Zero Dependencies)
Chỉ cần mở thư mục `fe` và **click đúp vào file `index.html`** để mở trên bất kỳ trình duyệt nào.

### Cách 2: Chạy qua HTTP Server tĩnh (Khuyên dùng)
```powershell
cd fe
python -m http.server 3000
```
Sau đó truy cập: **`http://localhost:3000`**
