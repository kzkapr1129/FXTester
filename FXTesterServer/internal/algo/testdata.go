package algo

// 日経225 (週足): 2023/3/17〜2024/4/19
var TestDataNikkei225Week = []Candle{
	{
		Label: "2023/03/17",
		Open:  27886.21,
		High:  27906.97,
		Low:   26632.92,
		Close: 27333.79,
	},
	{
		Label: "2023/03/24",
		Open:  27253.73,
		High:  27520.97,
		Low:   26945.67,
		Close: 27385.25,
	},
	{
		Label: "2023/03/31",
		Open:  27482.39,
		High:  28124.64,
		Low:   27359.72,
		Close: 28041.48,
	},
	{
		Label: "2023/04/07",
		Open:  28203.35,
		High:  28287.42,
		Low:   27427.66,
		Close: 27518.31,
	},
	{
		Label: "2023/04/14",
		Open:  27658.52,
		High:  28515.51,
		Low:   27597.18,
		Close: 28493.47,
	},
	{
		Label: "2023/04/21",
		Open:  28537.99,
		High:  28778.37,
		Low:   28414.98,
		Close: 28564.37,
	},

	{
		Label: "2023/04/28",
		Open:  28631.78,
		High:  28879.24,
		Low:   28241.67,
		Close: 28856.44,
	},

	{
		Label: "2023/05/02",
		Open:  29058.05,
		High:  29278.80,
		Low:   29016.83,
		Close: 29157.95,
	},

	{
		Label: "2023/05/12",
		Open:  29095.46,
		High:  29426.06,
		Low:   28931.81,
		Close: 29388.30,
	},

	{
		Label: "2023/05/19",
		Open:  29547.04,
		High:  30924.57,
		Low:   29475.97,
		Close: 30808.35,
	},

	{
		Label: "2023/05/26",
		Open:  30735.71,
		High:  31352.53,
		Low:   30558.14,
		Close: 30916.31,
	},

	{
		Label: "2023/06/02",
		Open:  31388.01,
		High:  31560.43,
		Low:   30785.98,
		Close: 31524.22,
	},

	{
		Label: "2023/06/09",
		Open:  31864.12,
		High:  32708.53,
		Low:   31420.45,
		Close: 32265.17,
	},

	{
		Label: "2023/06/16",
		Open:  32412.12,
		High:  33772.76,
		Low:   32280.54,
		Close: 33706.08,
	},

	{
		Label: "2023/06/23",
		Open:  33768.69,
		High:  33772.89,
		Low:   32575.56,
		Close: 32781.54,
	},

	{
		Label: "2023/06/30",
		Open:  32647.08,
		High:  33527.98,
		Low:   32306.99,
		Close: 33189.04,
	},

	{
		Label: "2023/07/07",
		Open:  33517.60,
		High:  33762.81,
		Low:   32327.90,
		Close: 32388.42,
	},

	{
		Label: "2023/07/14",
		Open:  32393.46,
		High:  32780.63,
		Low:   31791.71,
		Close: 32391.26,
	},

	{
		Label: "2023/07/21",
		Open:  32457.18,
		High:  32896.03,
		Low:   32080.95,
		Close: 32304.25,
	},

	{
		Label: "2023/07/28",
		Open:  32648.14,
		High:  32938.59,
		Low:   32037.55,
		Close: 32759.23,
	},

	{
		Label: "2023/08/04",
		Open:  33128.83,
		High:  33488.77,
		Low:   31934.35,
		Close: 32192.75,
	},

	{
		Label: "2023/08/10",
		Open:  31921.28,
		High:  32539.88,
		Low:   31830.23,
		Close: 32473.65,
	},

	{
		Label: "2023/08/18",
		Open:  32456.72,
		High:  32613.99,
		Low:   31275.25,
		Close: 31450.76,
	},

	{
		Label: "2023/08/25",
		Open:  31552.85,
		High:  32297.91,
		Low:   31409.86,
		Close: 31624.28,
	},

	{
		Label: "2023/09/01",
		Open:  31915.68,
		High:  32845.46,
		Low:   31881.93,
		Close: 32710.62,
	},

	{
		Label: "2023/09/08",
		Open:  32797.32,
		High:  33322.45,
		Low:   32512.80,
		Close: 32606.84,
	},

	{
		Label: "2023/09/15",
		Open:  32690.54,
		High:  33634.31,
		Low:   32391.69,
		Close: 33533.09,
	},

	{
		Label: "2023/09/22",
		Open:  33296.23,
		High:  33337.23,
		Low:   32154.53,
		Close: 32402.41,
	},

	{
		Label: "2023/09/29",
		Open:  32517.26,
		High:  32722.22,
		Low:   31674.42,
		Close: 31857.62,
	},

	{
		Label: "2023/10/06",
		Open:  32101.97,
		High:  32401.58,
		Low:   30487.67,
		Close: 30994.67,
	},

	{
		Label: "2023/10/13",
		Open:  31314.67,
		High:  32533.08,
		Low:   31314.67,
		Close: 32315.99,
	},

	{
		Label: "2023/10/20",
		Open:  31983.04,
		High:  32260.77,
		Low:   31093.90,
		Close: 31259.36,
	},

	{
		Label: "2023/10/27",
		Open:  31151.98,
		High:  31466.92,
		Low:   30551.67,
		Close: 30991.69,
	},

	{
		Label: "2023/11/02",
		Open:  30663.48,
		High:  32087.13,
		Low:   30538.29,
		Close: 31949.89,
	},

	{
		Label: "2023/11/10",
		Open:  32450.82,
		High:  32766.54,
		Low:   32049.34,
		Close: 32568.11,
	},

	{
		Label: "2023/11/17",
		Open:  32818.15,
		High:  33614.13,
		Low:   32499.28,
		Close: 33585.20,
	},

	{
		Label: "2023/11/24",
		Open:  33559.62,
		High:  33853.46,
		Low:   33182.99,
		Close: 33625.53,
	},

	{
		Label: "2023/12/01",
		Open:  33710.03,
		High:  33811.41,
		Low:   33161.07,
		Close: 33431.51,
	},

	{
		Label: "2023/12/08",
		Open:  33318.07,
		High:  33452.13,
		Low:   32205.38,
		Close: 32307.86,
	},

	{
		Label: "2023/12/15",
		Open:  32665.09,
		High:  33172.13,
		Low:   32515.04,
		Close: 32970.55,
	},

	{
		Label: "2023/12/22",
		Open:  32769.23,
		High:  33824.06,
		Low:   32541.23,
		Close: 33169.05,
	},

	{
		Label: "2023/12/29",
		Open:  33414.51,
		High:  33755.75,
		Low:   33181.36,
		Close: 33464.17,
	},

	{
		Label: "2024/01/05",
		Open:  33193.05,
		High:  33568.04,
		Low:   32693.18,
		Close: 33377.42,
	},

	{
		Label: "2024/01/12",
		Open:  33704.83,
		High:  35839.65,
		Low:   33600.32,
		Close: 35577.11,
	},

	{
		Label: "2024/01/19",
		Open:  35634.12,
		High:  36239.22,
		Low:   35371.25,
		Close: 35963.27,
	},

	{
		Label: "2024/02/02",
		Open:  35814.29,
		High:  36441.09,
		Low:   35704.58,
		Close: 36158.02,
	},

	{
		Label: "2024/02/09",
		Open:  36419.34,
		High:  37287.26,
		Low:   35854.63,
		Close: 36897.42,
	},

	{
		Label: "2024/02/16",
		Open:  37248.36,
		High:  38865.06,
		Low:   37184.10,
		Close: 38487.24,
	},

	{
		Label: "2024/02/22",
		Open:  38473.41,
		High:  39156.97,
		Low:   38095.15,
		Close: 39098.68,
	},

	{
		Label: "2024/03/01",
		Open:  39320.64,
		High:  39990.23,
		Low:   38876.81,
		Close: 39910.82,
	},

	{
		Label: "2024/03/08",
		Open:  40201.76,
		High:  40472.11,
		Low:   39518.40,
		Close: 39688.94,
	},

	{
		Label: "2024/03/15",
		Open:  39232.14,
		High:  39241.28,
		Low:   38271.38,
		Close: 38707.64,
	},

	{
		Label: "2024/03/22",
		Open:  38960.99,
		High:  41087.75,
		Low:   38935.47,
		Close: 40888.43,
	},

	{
		Label: "2024/03/29",
		Open:  40798.96,
		High:  40979.36,
		Low:   40054.06,
		Close: 40369.44,
	},

	{
		Label: "2024/04/05",
		Open:  40646.70,
		High:  40697.22,
		Low:   38774.24,
		Close: 38992.08,
	},

	{
		Label: "2024/04/12",
		Open:  39391.98,
		High:  39774.82,
		Low:   39065.31,
		Close: 39523.55,
	},

	{
		Label: "2024/04/19",
		Open:  39056.93,
		High:  39232.80,
		Low:   36733.06,
		Close: 37068.35,
	},

	{
		Label: "2024/04/26",
		Open:  37240.93,
		High:  38460.08,
		Low:   37052.63,
		Close: 37934.76,
	},

	{
		Label: "2024/05/02",
		Open:  38312.66,
		High:  38608.17,
		Low:   37958.19,
		Close: 38236.07,
	},

	{
		Label: "2024/05/10",
		Open:  38636.23,
		High:  38863.14,
		Low:   38072.24,
		Close: 38229.11,
	},

	{
		Label: "2024/05/17",
		Open:  38211.61,
		High:  38949.38,
		Low:   37969.58,
		Close: 38787.38,
	},

	{
		Label: "2024/05/24",
		Open:  38761.71,
		High:  39437.16,
		Low:   38367.70,
		Close: 38646.11,
	},

	{
		Label: "2024/05/31",
		Open:  38766.21,
		High:  39141.99,
		Low:   37617.00,
		Close: 38487.90,
	},

	{
		Label: "2024/06/07",
		Open:  38734.95,
		High:  39032.50,
		Low:   38343.98,
		Close: 38683.93,
	},

	{
		Label: "2024/06/14",
		Open:  38689.78,
		High:  39336.66,
		Low:   38554.75,
		Close: 38814.56,
	},

	{
		Label: "2024/06/21",
		Open:  38440.98,
		High:  38797.97,
		Low:   37950.20,
		Close: 38596.47,
	},

	{
		Label: "2024/06/28",
		Open:  38497.42,
		High:  39788.63,
		Low:   38416.07,
		Close: 39583.08,
	},

	{
		Label: "2024/07/05",
		Open:  39839.82,
		High:  41100.13,
		Low:   39457.62,
		Close: 40912.37,
	},

	{
		Label: "2024/07/12",
		Open:  40863.14,
		High:  42426.77,
		Low:   40780.70,
		Close: 41190.68,
	},

	{
		Label: "2024/07/19",
		Open:  41366.79,
		High:  41520.07,
		Low:   39824.58,
		Close: 40063.79,
	},

	{
		Label: "2024/07/26",
		Open:  39947.95,
		High:  39973.20,
		Low:   37611.19,
		Close: 37667.41,
	},

	{
		Label: "2024/08/02",
		Open:  38139.12,
		High:  39188.37,
		Low:   35880.15,
		Close: 35909.70,
	},

	{
		Label: "2024/08/09",
		Open:  35249.36,
		High:  35849.77,
		Low:   31156.12,
		Close: 35025.00,
	},

	{
		Label: "2024/08/16",
		Open:  35490.58,
		High:  38143.55,
		Low:   35476.79,
		Close: 38062.67,
	},

	{
		Label: "2024/08/23",
		Open:  37863.76,
		High:  38424.27,
		Low:   37318.04,
		Close: 38364.27,
	},

	{
		Label: "2024/08/30",
		Open:  38156.41,
		High:  38669.79,
		Low:   37825.31,
		Close: 38647.75,
	},

	{
		Label: "2024/09/06",
		Open:  39025.31,
		High:  39080.64,
		Low:   36235.61,
		Close: 36391.47,
	},
}

var TestDataNikkei225Week_Result_peaks = []ZigzagResult{
	{
		Info:        "2023/03/17",
		PeakIndex:   0,
		BottomIndex: 1,
	},
	{
		Info:        "2023/06/23",
		PeakIndex:   14,
		BottomIndex: 18,
	},
	{
		Info:        "2023/08/04",
		PeakIndex:   20,
		BottomIndex: 22,
	},
	{
		Info:        "2023/09/15",
		PeakIndex:   26,
		BottomIndex: 29,
	},
	{
		Info:        "2023/10/13",
		PeakIndex:   30,
		BottomIndex: 33,
	},
	{
		Info:        "2023/12/01",
		PeakIndex:   37,
		BottomIndex: 38,
	},
	{
		Info:        "2023/12/29",
		PeakIndex:   41,
		BottomIndex: 42,
	},
	{
		Info:        "2024/03/08",
		PeakIndex:   50,
		BottomIndex: 51,
	},
	{
		Info:        "2024/03/22",
		PeakIndex:   52,
		BottomIndex: 56,
	},
	{
		Info:        "2024/05/17",
		PeakIndex:   60,
		BottomIndex: 62,
	},
	{
		Info:        "2024/06/14",
		PeakIndex:   64,
		BottomIndex: 65,
	},
	{
		Info:        "2024/07/19",
		PeakIndex:   69,
		BottomIndex: 72,
	},
}

var TestDataNikkei225Week_Result_bottoms = []ZigzagResult{
	{
		Info:        "2023/03/24",
		PeakIndex:   14,
		BottomIndex: 1,
	},

	{
		Info:        "2023/07/21",
		PeakIndex:   20,
		BottomIndex: 18,
	},

	{
		Info:        "2023/08/18",
		PeakIndex:   26,
		BottomIndex: 22,
	},

	{
		Info:        "2023/10/06",
		PeakIndex:   30,
		BottomIndex: 29,
	},

	{
		Info:        "2023/11/02",
		PeakIndex:   37,
		BottomIndex: 33,
	},

	{
		Info:        "2023/12/08",
		PeakIndex:   41,
		BottomIndex: 38,
	},

	{
		Info:        "2024/01/05",
		PeakIndex:   50,
		BottomIndex: 42,
	},

	{
		Info:        "2024/03/15",
		PeakIndex:   52,
		BottomIndex: 51,
	},

	{
		Info:        "2024/04/19",
		PeakIndex:   60,
		BottomIndex: 56,
	},

	{
		Info:        "2024/05/31",
		PeakIndex:   64,
		BottomIndex: 62,
	},

	{
		Info:        "2024/06/21",
		PeakIndex:   69,
		BottomIndex: 65,
	},

	{
		Info:        "2024/08/09",
		PeakIndex:   76,
		BottomIndex: 72,
	},
}
