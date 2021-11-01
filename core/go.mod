module github.com/kurtosis-tech/kurtosis-core

go 1.15

require (
    github.com/kurtosis-tech/kurtosis-core/api/golang v0.0.0
    github.com/kurtosis-tech/kurtosis-core/server v0.0.0
)

replace (
    github.com/kurtosis-tech/kurtosis-core/api/golang => ./api/golang
    github.com/kurtosis-tech/kurtosis-core/server => ./server
)

