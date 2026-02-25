import os
import csv
import librosa
import numpy as np

TRAIN_DIR = "Data_Train_to_DB"
OUTPUT_CSV = "audio_features.csv"

# Vector đặc trưng 19 chiều:
# - 13 MFCC mean      → Âm sắc (timbre), phân biệt các nhạc cụ khác nhau
# - 1 Spectral Centroid → Độ sáng phổ (brightness), tần số trung tâm
# - 1 Spectral Bandwidth → Độ rộng phổ, dải tần hoạt động của nhạc cụ
# - 1 Spectral Rolloff  → Tần số mà 85% năng lượng nằm dưới
# - 1 Zero Crossing Rate → Tỉ lệ tín hiệu đổi dấu, phân biệt âm trầm/cao
# - 1 RMS Energy        → Độ to (loudness), năng lượng trung bình
# - 1 Spectral Contrast → Tương phản phổ, phân biệt harmonic vs noise

FEATURE_NAMES = (
    [f"mfcc_{i + 1}" for i in range(13)]
    +
    [
        "spectral_centroid",
        "spectral_bandwidth",
        "spectral_rolloff",
        "zero_crossing_rate",
        "rms_energy",
        "spectral_contrast"
    ]
)


def extract_features(file_path):
    """
        Đọc file .wav và trích xuất vector đặc trưng 19 chiều.
        Sử dụng sample rate 22050 Hz (chuẩn cho phân tích âm nhạc).
    """
    y , sr = librosa.load(file_path,sr = 22050)
    # 1. MFCC - 13 hệ số (đặc trưng âm sắc chính)
    mfcc = librosa.feature.mfcc(y=y,sr=sr,n_mfcc=13)
    mfcc_mean = np.mean(mfcc,axis=1)
    # 2. Spectral Centroid (Brightness - Độ sáng phổ)
    centroid = librosa.feature.spectral_centroid(y=y,sr=sr)
    centroid_mean = np.mean(centroid)
    # 3. Spectral Bandwidth (Độ rộng phổ)
    bandwidth = librosa.feature.spectral_bandwidth(y=y, sr=sr)
    bandwidth_mean = np.mean(bandwidth)
    # 4. Spectral Rolloff (Tần số rolloff)
    rolloff = librosa.feature.spectral_rolloff(y=y, sr=sr)
    rolloff_mean = np.mean(rolloff)

    # 5. Zero Crossing Rate (Tỉ lệ cắt zero)
    zcr = librosa.feature.zero_crossing_rate(y)
    zcr_mean = np.mean(zcr)
    # 6. RMS Energy (Loudness - Độ to)
    rms = librosa.feature.rms(y=y)
    rms_mean = np.mean(rms)
    # 7. Spectral Contrast (Tương phản phổ)
    contrast = librosa.feature.spectral_contrast(y=y, sr=sr)
    contrast_mean = np.mean(contrast)
    # Ghép thành vector 19 chiều
    feature_vector = np.concatenate([
        mfcc_mean,  # 13 dims
        [centroid_mean],  # 1 dim
        [bandwidth_mean],  # 1 dim
        [rolloff_mean],  # 1 dim
        [zcr_mean],  # 1 dim
        [rms_mean],  # 1 dim
        [contrast_mean]  # 1 dim
    ])
    return feature_vector

def main():
    header = FEATURE_NAMES + ["instrument", "filename"]

    processed = 0
    errors = 0
    with open(OUTPUT_CSV, 'w', newline='', encoding='utf-8') as f:
        writer = csv.writer(f)
        writer.writerow(header)  # Ghi dòng tiêu đề đầu tiên

        instruments = sorted([
            d for d in os.listdir(TRAIN_DIR)
            if os.path.isdir(os.path.join(TRAIN_DIR, d))
        ]) # Lấy thư mục con và sắp xếp A -> Z

        for instrument in instruments:
            inst_path = os.path.join(TRAIN_DIR, instrument)
            wav_files = sorted([f for f in os.listdir(inst_path) if f.endswith('.wav')]) # ấy tất cả file .wav trong đó, sắp xếp A→Z
            print(f"\n🎵 {instrument} ({len(wav_files)} files):")
            for i , filename in enumerate(wav_files):
                filepath = os.path.join(inst_path, filename)
                try:
                    features = extract_features(filepath)  #  Trích xuất 19 đặc trưng
                    row = [f"{v:.6f}" for v in features] + [instrument, filename]  #  Tạo dòng CSV
                    writer.writerow(row)  #  Ghi vào file
                    processed += 1
                    print(f"  [{i + 1}/{len(wav_files)}] {filename}")
                except Exception as e:
                    errors += 1
                    print(f"   [{i + 1}/{len(wav_files)}] {filename} → Lỗi: {e}")

    print(f"\n{'=' * 60}")
    print(f"🎉 Hoàn thành trích xuất đặc trưng!")
    print(f"   ✅ Thành công: {processed} file")
    print(f"   ❌ Lỗi:       {errors} file")
    print(f"   📄 Kết quả:   '{OUTPUT_CSV}'")
    print(f"   📐 Vector:    {len(FEATURE_NAMES)} chiều")
    print(f"{'=' * 60}")

if __name__ == "__main__":
    main()