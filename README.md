# gost

[![Build Status](https://cloud.drone.io/api/badges/ch1aki/gost/status.svg)](https://cloud.drone.io/ch1aki/gost)

A CLI command to simply compute statistics.

## Installing

```
brew tap ch1aki/homebrew-gost
brew install gost
```

or

```
$ go get -u github.com/ch1aki/gost
```

## Usage

```
$ seq 1 100 | gost
N       min     max     sum     mean    stddev
100     1       100     5050    50.5    29.011491975882016
```

## Authors

ch1aki

## License

Licensed under the MIT License.
