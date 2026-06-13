import sys
import os
import csv
import librosa
import numpy as np
import warnings
warnings.filterwarnings('ignore') # Ẩn các cảnh báo rác của librosa

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


def load_scaler_params(scaler_csv_path):
    """
        Đọc file scaler_params.csv (mean, std của 19 đặc trưng tính từ tập train)
        và trả về dict {tên_đặc_trưng: (mean, std)} để dùng cho việc chuẩn hóa.
    """
    if not scaler_csv_path or not os.path.exists(scaler_csv_path):
        return {}

    params = {}
    with open(scaler_csv_path, 'r', newline='', encoding='utf-8') as f:
        reader = csv.reader(f)
        # Skip header
        next(reader, None)
        for row in reader:
            if len(row) < 3:
                continue
            feature_name = row[0]
            try:
                mean_val = float(row[1])
                std_val = float(row[2])
            except ValueError:
                continue
            params[feature_name] = (mean_val, std_val)
    return params


def apply_scaler(feature_vector, scaler_params):
    """
        Chuẩn hóa Z-score vector đặc trưng: (x - mean) / std,
        dùng đúng mean/std đã tính từ tập train để đảm bảo cùng thang đo với DB.
    """
    if not scaler_params:
        return feature_vector

    scaled = feature_vector.copy()
    for i, name in enumerate(FEATURE_NAMES):
        if name not in scaler_params:
            continue
        mean_val, std_val = scaler_params[name]
        if std_val == 0:
            continue
        scaled[i] = (scaled[i] - mean_val) / std_val
    return scaled

def extract_single_feature(file_path, scaler_csv_path=None):
    """
        Trích xuất vector 19 chiều cho 1 file audio (giống công thức extract_features.py),
        chuẩn hóa bằng scaler_params đã lưu, rồi in ra stdout để Go đọc kết quả.
    """
    try:
        y, sr = librosa.load(file_path, sr=22050)
        # Trích xuất dữ liệu đặc trưng 
        mfcc_mean = np.mean(librosa.feature.mfcc(y=y, sr=sr, n_mfcc=13), axis=1)
        centroid_mean = np.mean(librosa.feature.spectral_centroid(y=y, sr=sr))
        bandwidth_mean = np.mean(librosa.feature.spectral_bandwidth(y=y, sr=sr))
        rolloff_mean = np.mean(librosa.feature.spectral_rolloff(y=y, sr=sr))
        zcr_mean = np.mean(librosa.feature.zero_crossing_rate(y))
        rms_mean = np.mean(librosa.feature.rms(y=y))
        contrast_mean = np.mean(librosa.feature.spectral_contrast(y=y, sr=sr))

        # Gom thành 1 mảng
        feature_vector = np.concatenate([
            mfcc_mean, [centroid_mean], [bandwidth_mean], 
            [rolloff_mean], [zcr_mean], [rms_mean], [contrast_mean]
        ])

        scaler_params = load_scaler_params(scaler_csv_path)
        feature_vector = apply_scaler(feature_vector, scaler_params)

        # In ra màn hình các số cách nhau bởi dấu cách để Golang đọc
        print(" ".join([f"{v:.6f}" for v in feature_vector]))
    except Exception as e:
        print(f"ERROR: {e}", file=sys.stderr)
        sys.exit(1)    
if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("ERROR: No file path provided", file=sys.stderr)
        sys.exit(1)
    scaler_csv = None
    if len(sys.argv) >= 3:
        scaler_csv = sys.argv[2]
    else:
        scaler_csv = "../audio-feature-extractor/scaler_params.csv"
    extract_single_feature(sys.argv[1], scaler_csv)