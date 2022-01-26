package metrics_optin

const (
	WhyKurtosisCollectMetricsDescriptionNote = "NOTE: I know that user metrics are heavily abused in today's world to invade users' \n" +
		"privacy, and I hate it too. This is my philosophy: metrics are very useful for \n" +
		"detecting product bugs and seeing the most popular features so we can invest more in \n" +
		"them, but they should *only* be used for that. It was important to me that our metrics be:\n" +
		"   a) opt-in (we require you to make a choice about collection rather than assuming)\n" +
		"   b) anonymized (your user ID is a hash; we don't know who you are)\n" +
		"   c) obfuscated (your parameters are hashed as well)\n" +
		"   d) private (we will never give or sell your data to third parties)\n" +
		"You opting in to metrics will improve the product for everyone, so it's important to me that\n" +
		"we honor the trust you place in us by doing the above to keep your privacy sacred.\n" +
		"                                                                       Kevin Today, CTO, Kurtosis Technologies"
)