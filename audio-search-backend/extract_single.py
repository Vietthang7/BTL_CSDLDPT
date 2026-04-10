import sys
import librosa
import numpy as np
import warnings
warnings.filterwarnings('ignore') # Ẩn các cảnh báo rác của librosa

def extract_single_feature(file_path):
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

        # In ra màn hình các số cách nhau bởi dấu cách để Golang đọc
        print(" ".join([f"{v:.6f}" for v in feature_vector]))
    except Exception as e:
        print(f"ERROR: {e}", file=sys.stderr)
        sys.exit(1)    
if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("ERROR: No file path provided", file=sys.stderr)
        sys.exit(1)
    extract_single_feature(sys.argv[1])