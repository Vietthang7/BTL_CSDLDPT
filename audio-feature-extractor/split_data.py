import os
import shutil
import random

# Cấu hình
SOURCE_ROOT = "Dataset_BTL_600_Final"
TRAIN_DIR = "Data_Train_to_DB"
TEST_DIR = "Data_Test"
TEST_LIMIT_PER_TYPE = 20  # Mỗi loại lấy 20 file làm test


def split_data():
    # Tạo thư mục mới
    for d in [TRAIN_DIR, TEST_DIR]:
        if os.path.exists(d): shutil.rmtree(d)
        os.makedirs(d)

    instruments = [d for d in os.listdir(SOURCE_ROOT) if os.path.isdir(os.path.join(SOURCE_ROOT, d))]

    for inst in instruments:
        inst_path = os.path.join(SOURCE_ROOT, inst)
        files = [f for f in os.listdir(inst_path) if f.endswith('.wav')]

        # Trộn ngẫu nhiên để tập test khách quan
        random.shuffle(files)

        # Chia danh sách
        test_files = files[:TEST_LIMIT_PER_TYPE]
        train_files = files[TEST_LIMIT_PER_TYPE:]

        # Tạo thư mục con
        os.makedirs(os.path.join(TEST_DIR, inst))
        os.makedirs(os.path.join(TRAIN_DIR, inst))

        # Copy vào thư mục Test
        for f in test_files:
            shutil.copy2(os.path.join(inst_path, f), os.path.join(TEST_DIR, inst, f))

        # Copy vào thư mục Train
        for f in train_files:
            shutil.copy2(os.path.join(inst_path, f), os.path.join(TRAIN_DIR, inst, f))

        print(f"--- {inst}: Đã tách 20 file vào Test và {len(train_files)} file vào Train.")

    print(f"\n✅ Hoàn thành!")
    print(f"- Thư mục '{TRAIN_DIR}': Dùng để trích xuất đặc trưng và nạp DB (520 file).")
    print(f"- Thư mục '{TEST_DIR}': Dùng để upload lên web demo thử (80 file).")


if __name__ == "__main__":
    split_data()