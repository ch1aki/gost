# gost

[![Build Status](https://cloud.drone.io/api/badges/ch1aki/gost/status.svg)](https://cloud.drone.io/ch1aki/gost)

A CLI command to simply compute statistics.

## Installing

```
brew tap ch1aki/homebrew-gost
brew install ch1aki/gost/gost
```

or

```
$ go get -u github.com/ch1aki/gost
```

## Usage

```
Usage:
  gost [OPTIONS]

Application Options:
  -v, --version   show version
  -f, --format=   outpuut format (default: %g)
  -N, --count     sample size
      --min       minimum
      --max       maximum
      --sum       sum of elements of the sample
  -m, --mean      mean
      --sd        standard deviation
      --variance  variance
      --markdown  markdown table format

Help Options:
  -h, --help      Show this help message
```

```
$ seq 1 100 | gost
N       MIN     MAX     SUM     MEAN    STDDEV             
100     1       100     5050    50.5    29.011491975882016

$ seq 1 100 | gost --markdown
|  N  | MIN | MAX | SUM  | MEAN |       STDDEV       |
|-----|-----|-----|------|------|--------------------|
| 100 |   1 | 100 | 5050 | 50.5 | 29.011491975882016 |
```

## Authors

ch1aki

## License

Licensed under the MIT License.
