package validator

import (
	"fxtester/internal/gen"
	"testing"
)

func Test_ErrorHandler(t *testing.T) {
	type args struct {
		candle gen.Candle
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "正常ケース1",
			args: args{
				candle: gen.Candle{
					Time:  "2024-01-01T00:00:00Z",
					High:  120.00,
					Open:  110.00,
					Close: 100.00,
					Low:   90.00,
				},
			},
		},
		{
			name: "正常ケース2",
			args: args{
				candle: gen.Candle{
					Time:  "2024-12-31T23:59:59.999+23:59",
					High:  120.00,
					Open:  110.00,
					Close: 100.00,
					Low:   90.00,
				},
			},
		},
		{
			name: "日時がISO8601以外のフォーマット",
			args: args{
				candle: gen.Candle{
					Time:  "2024/12/31 23:59:59.999",
					High:  0,
					Open:  0,
					Close: 0,
					Low:   0,
				},
			},
			wantErr: true,
		},
		{
			name: "高値がマイナス",
			args: args{
				candle: gen.Candle{
					Time:  "2024-12-31T23:59:59.999+23:59",
					High:  -1,
					Open:  0,
					Close: 0,
					Low:   0,
				},
			},
			wantErr: true,
		},
		{
			name: "始値がマイナス",
			args: args{
				candle: gen.Candle{
					Time:  "2024-12-31T23:59:59.999+23:59",
					High:  0,
					Open:  -1,
					Close: 0,
					Low:   0,
				},
			},
			wantErr: true,
		},
		{
			name: "終値がマイナス",
			args: args{
				candle: gen.Candle{
					Time:  "2024-12-31T23:59:59.999+23:59",
					High:  0,
					Open:  0,
					Close: -1,
					Low:   0,
				},
			},
			wantErr: true,
		},
		{
			name: "安値がマイナス",
			args: args{
				candle: gen.Candle{
					Time:  "2024-12-31T23:59:59.999+23:59",
					High:  0,
					Open:  0,
					Close: 0,
					Low:   -1,
				},
			},
			wantErr: true,
		},
		{
			name: "高値が始値より低い",
			args: args{
				candle: gen.Candle{
					Time:  "2024-12-31T23:59:59.999+23:59",
					High:  2,
					Open:  4,
					Close: 1,
					Low:   1,
				},
			},
			wantErr: true,
		},
		{
			name: "高値が終値より低い",
			args: args{
				candle: gen.Candle{
					Time:  "2024-12-31T23:59:59.999+23:59",
					High:  4,
					Open:  4,
					Close: 5,
					Low:   1,
				},
			},
			wantErr: true,
		},
		{
			name: "高値が安値より低い",
			args: args{
				candle: gen.Candle{
					Time:  "2024-12-31T23:59:59.999+23:59",
					High:  3,
					Open:  4,
					Close: 4,
					Low:   4,
				},
			},
			wantErr: true,
		},
		{
			name: "安値が始値より高い",
			args: args{
				candle: gen.Candle{
					Time:  "2024-12-31T23:59:59.999+23:59",
					High:  6,
					Open:  4,
					Close: 5,
					Low:   5,
				},
			},
			wantErr: true,
		},
		{
			name: "安値が終値より高い",
			args: args{
				candle: gen.Candle{
					Time:  "2024-12-31T23:59:59.999+23:59",
					High:  7,
					Open:  6,
					Close: 4,
					Low:   5,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := ValidateCandle(tt.args.candle); (err != nil) != tt.wantErr {
				t.Errorf("ValidateCandle()=%v wantErr=%v", err, tt.wantErr)
			}
		})
	}
}
