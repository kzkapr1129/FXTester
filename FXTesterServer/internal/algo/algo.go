package algo

import (
	"errors"
	"math"
)

type Candle struct {
	Label string
	High  float64
	Open  float64
	Close float64
	Low   float64
}

type Kind int

const (
	Peak Kind = iota
	Bottom
)

type ZigzagResult struct {
	Info        string
	PeakIndex   int
	BottomIndex int
	Velocity    float64
	Delta       float64
	Kind        Kind
}

func (c *Candle) BoxMax() float64 {
	return math.Max(c.Open, c.Close)
}

func (c *Candle) BoxMin() float64 {
	return math.Min(c.Open, c.Close)
}

func (c *Candle) Contains(t *Candle) bool {
	return c.BoxMin() <= t.BoxMin() && t.BoxMax() <= c.BoxMax()
}

func (c *Candle) isUpdatedBoxMaxBy(t *Candle) bool {
	return c.BoxMax() < t.BoxMax()
}

func (c *Candle) isUpdatedBoxMinBy(t *Candle) bool {
	return t.BoxMin() < c.BoxMin()
}

func (c *Candle) isPositive() bool {
	return c.Open < c.Close
}

func (c *Candle) isNegative() bool {
	return c.Close < c.Open
}

func FindZigzagPeak(candles []Candle) []ZigzagResult {
	results := make([]ZigzagResult, 0)

	for i := 0; i < len(candles); i++ {

		// 高値更新が止まった場所を探す
		peakIndex, bottomStart, peakCandle, err := findPeak(candles, i)
		if err != nil {
			// プログラムのミス
			panic(err)
		}

		// 安値更新が止まった場所を探す
		bottomIndex, _, bottomCandle, err := findBottom(candles, bottomStart)
		if err != nil {
			// プログラムのミス
			panic(err)
		}

		if bottomIndex <= peakIndex {
			/*
			 * グラフが高値更新しかしていない場合、`peakIndex == bottomIndex`となる可能性がある。
			 * この場合は検出なしとして処理する
			 */
			break
		}

		// 経過時間
		x := bottomIndex - peakIndex
		// Y軸のΔ
		y := peakCandle.BoxMax() - bottomCandle.BoxMin()
		// 速度を計算
		velocity := y / float64(x)

		results = append(results, ZigzagResult{
			Info:        candles[peakIndex].Label,
			PeakIndex:   peakIndex,
			BottomIndex: bottomIndex,
			Velocity:    velocity,
			Delta:       y,
			Kind:        Peak,
		})

		// 同じ箇所の判定を避けるため、検査済みのインデックスまで進める
		i = bottomIndex
	}

	// slices.SortFunc(results, func(i, j ZigzagResult) int {
	// 	if i.Velocity == j.Velocity {
	// 		return 0
	// 	}
	// 	return func() int {
	// 		if i.Velocity > j.Velocity {
	// 			return -1
	// 		} else {
	// 			return 1
	// 		}
	// 	}()
	// })

	return results
}

/*
 * findPeak 最も高値更新したローソク足を探す。ネックライン割れ、安値更新が起きた場合は高値更新は終了したと判断する。
 */
func findPeak(candles []Candle, start int) (int, int, *Candle, error) {
	if len(candles) <= start {
		return 0, 0, nil, errors.New("overflow")
	}

	peakIndex := start
	if len(candles) <= peakIndex+1 {
		return peakIndex, peakIndex, &candles[peakIndex], nil
	}

	peak := candles[peakIndex]
	lastIndex := peakIndex

	// 高値更新が続いている間繰り返す
	for i := peakIndex + 1; i < len(candles); i++ {
		lastIndex = i // 処理済みのインデックスを保持しておく

		c := candles[i]
		if peak.isUpdatedBoxMaxBy(&c) {
			// 高値更新があった場合
			peak = c
			peakIndex = i
			continue
		}

		// ローソク足の包含関係を確認
		prev := candles[i-1]
		if prev.Contains(&c) {
			// 前回のローソク足に包含されている場合
			continue
		} else if prev.isUpdatedBoxMaxBy(&c) && !prev.isUpdatedBoxMinBy(&c) {
			// 前回のローソク足の高値を更新した場合 (peakの更新はなし、安値の更新はなし)
			continue
		} else if prev.isUpdatedBoxMinBy(&c) {
			// 前回のローソク足の安値を更新した場合
			if prev.isUpdatedBoxMaxBy(&c) {
				// 安値と高値(peakの更新はなし)両方更新した場合

				if c.isPositive() {
					// ローソク足が陽線の場合
					continue // 高値更新を優先する (処理継続)
				} else if c.isNegative() {
					// ローソク足が陰線の場合
					break // 安値更新を優先する (処理終了)
				} else {
					// 十字線の場合

					// プログラムのミスまたは検討不足な問題(包含関係ではないので十字線はあり得ない)
					panic("unexpected case")
				}
			} else {
				// 安値だけ更新した場合
				break
			}
		} else {
			// プログラムのミスまたは検討不足な問題(包含関係ではなく、高値・安値更新でもない)
			panic("unexpected case")
		}
	}

	return peakIndex, lastIndex, &peak, nil
}

/*
 * findPeak 最も安値更新したローソク足を探す。ネックライン割れ、高値更新が起きた場合は高値更新は終了したと判断する。
 */
func findBottom(candles []Candle, start int) (int, int, *Candle, error) {
	if len(candles) <= start {
		return 0, 0, nil, errors.New("overflow")
	}

	bottomIndex := start
	if len(candles) <= bottomIndex+1 {
		return bottomIndex, bottomIndex, &candles[bottomIndex], nil
	}

	bottom := candles[bottomIndex]
	lastIndex := bottomIndex

	// 安値更新が続いている間繰り返す
	for i := bottomIndex + 1; i < len(candles); i++ {
		lastIndex = i // 処理済みのインデックスを保持しておく

		c := candles[i]
		if bottom.isUpdatedBoxMinBy(&c) {
			// 安値更新があった場合
			bottom = c
			bottomIndex = i
			continue
		}

		// ローソク足の包含関係を確認
		prev := candles[i-1]
		if prev.Contains(&c) {
			// 前回のローソク足に包含されている場合
			continue
		} else if prev.isUpdatedBoxMinBy(&c) && !prev.isUpdatedBoxMaxBy(&c) {
			// 前回のローソク足の安値を更新した場合 (bottomの更新はなし、高値の更新はなし)
			continue
		} else if prev.isUpdatedBoxMaxBy(&c) {
			// 前回のローソク足の高値を更新した場合
			if prev.isUpdatedBoxMinBy(&c) {
				// 高値と安値(bottomの更新はなし)両方更新した場合

				if c.isNegative() {
					// ローソク足が陰線の場合
					continue // 安値更新を優先する (処理継続)
				} else if c.isPositive() {
					// ローソク足が陽線の場合
					break // 高値更新を優先する (処理終了)
				} else {
					// 十字線の場合

					// プログラムのミスまたは検討不足な問題(包含関係ではないので十字線はあり得ない)
					panic("unexpected case")
				}
			} else {
				// 高値だけ更新した場合
				break
			}
		} else {
			// プログラムのミスまたは検討不足な問題(包含関係ではなく、高値・安値更新でもない)
			panic("unexpected case")
		}
	}

	return bottomIndex, lastIndex, &bottom, nil
}
