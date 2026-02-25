import os
import shutil

SOURCE_DIR = "IRMAS-TrainingData"
TARGET_DIR = "Dataset_BTL_600_Final"
INSTRUMENTS = {"vio": "violin", "cel": "cello", "gac": "acoustic_guitar", "gel": "electric_guitar"}
LIMIT_PER_TYPE = 150

# Các tag nhạc cụ khác trong IRMAS cần loại bỏ để đảm bảo đơn tấu (solo)
OTHER_INSTRUMENT_TAGS = [
    "[dru]",  # drums
    "[voi]",  # voice
    "[pia]",  # piano
    "[org]",  # organ
    "[tru]",  # trumpet
    "[flu]",  # flute
    "[sax]",  # saxophone
    "[cla]",  # clarinet
]


def is_solo(filename):
    """Kiểm tra file có phải đơn tấu không (không lẫn nhạc cụ khác)"""
    for tag in OTHER_INSTRUMENT_TAGS:
        if tag in filename:
            return False
    return True


def prepare_data():
    if os.path.exists(TARGET_DIR):
        shutil.rmtree(TARGET_DIR)
    os.makedirs(TARGET_DIR)

    total_count = 0
    for code, folder_name in INSTRUMENTS.items():
        src_path = os.path.join(SOURCE_DIR, code)
        dest_path = os.path.join(TARGET_DIR, folder_name)
        if not os.path.exists(dest_path):
            os.makedirs(dest_path)

        all_files = [f for f in os.listdir(src_path) if f.endswith('.wav')]

        # Ưu tiên 1: [nod] + solo → Tốt nhất (không trống, không lẫn nhạc cụ khác)
        solo_nod = [f for f in all_files if "[nod]" in f and is_solo(f)]

        # Ưu tiên 2: [nod] + có lẫn nhạc cụ khác → Vẫn tốt vì không có trống
        nod_mixed = [f for f in all_files if "[nod]" in f and not is_solo(f)]

        # Ưu tiên 3: Không [nod] nhưng solo (không lẫn nhạc cụ nào kể cả drums)
        solo_no_nod = [f for f in all_files if "[nod]" not in f and is_solo(f)]

        # Ưu tiên 4: Còn lại, chỉ loại [dru] (backup cuối cùng)
        remaining = [f for f in all_files if f not in solo_nod and f not in nod_mixed
                     and f not in solo_no_nod and "[dru]" not in f]

        # Ghép theo thứ tự ưu tiên: [nod] LUÔN được ưu tiên trước
        final_list = (solo_nod + nod_mixed + solo_no_nod + remaining)[:LIMIT_PER_TYPE]

        print(f"--- {folder_name}:")
        print(f"    [nod] + solo:    {len(solo_nod)} file (tốt nhất)")
        print(f"    [nod] + mix:     {len(nod_mixed)} file (không trống, có đệm)")
        print(f"    Solo, không nod: {len(solo_no_nod)} file (có thể có trống)")
        print(f"    Backup:          {len(remaining)} file")
        print(f"    → Lấy: {len(final_list)} / {len(all_files)} file")

        for f in final_list:
            shutil.copy2(os.path.join(src_path, f), os.path.join(dest_path, f))
            total_count += 1

    print(f"\n✅ Tổng cộng: {total_count} file đơn tấu (solo) nhạc cụ bộ dây.")
    print(f"📁 Dữ liệu đã lưu tại: '{TARGET_DIR}'")
    if total_count >= 500:
        print(f"🎯 Đã đủ điều kiện ≥ 500 file của thầy!")
    else:
        print(f"⚠️  Mới có {total_count} file, cần thêm {500 - total_count} file nữa!")


if __name__ == "__main__":
    prepare_data()