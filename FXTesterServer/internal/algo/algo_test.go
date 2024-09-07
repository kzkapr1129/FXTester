package algo

import (
	"slices"
	"strings"
	"testing"
)

func Test_FindZigzagPeak(t *testing.T) {
	type args struct {
		input []Candle
	}

	tests := []struct {
		name        string
		args        args
		wantPanic   bool
		wantResults []ZigzagResult
	}{
		{
			name: "normal",
			args: args{
				input: TestDataNikkei225Week,
			},
			wantPanic:   false,
			wantResults: TestDataNikkei225Week_Result_peaks,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			defer func() {
				if e := recover(); (e != nil) != tt.wantPanic {
					t.Errorf("FindZigzagPeak()=%v wantPanic=%v", e, tt.wantPanic)
				} else if e != nil {
					panic(e)
				}
			}()

			checkResults := func(results []ZigzagResult) {
				if len(results) != len(tt.wantResults) {
					t.Errorf("FindZigzagPeak()=%v len(wantResults)=%v", len(results), len(tt.wantResults))
				}

				// ZigzagResultの存在チェック
				checkContains := func(src []ZigzagResult, dest []ZigzagResult, tag string) {
					notFounds := []string{}
					for _, result := range src {
						contains := slices.ContainsFunc(dest, func(z ZigzagResult) bool {
							return result.Info == z.Info
						})
						if !contains {
							notFounds = append(notFounds, result.Info)
						}
					}
					if 0 < len(notFounds) {
						t.Errorf("Couldn't find in %s: %s", tag, strings.Join(notFounds, ","))
					}
				}
				checkContains(results, tt.wantResults, "expects")
				checkContains(tt.wantResults, results, "results")

				// 結果の内容チェック
				for _, result := range results {
					expectIndex := slices.IndexFunc(tt.wantResults, func(v ZigzagResult) bool {
						return v.Info == result.Info
					})
					if expectIndex != -1 {
						expect := tt.wantResults[expectIndex]
						if result.PeakIndex != expect.PeakIndex || result.BottomIndex != expect.BottomIndex {
							t.Errorf("Couldn't match indexes: info=%s expect=(%d,%d) actual=(%d, %d)",
								result.Info,
								expect.PeakIndex, expect.BottomIndex,
								result.PeakIndex, result.BottomIndex)
						}
					}
				}

			}

			results := FindZigzagPeak(tt.args.input)
			checkResults(results)
		})
	}
}
