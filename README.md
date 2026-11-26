# gost

A CLI command to simply compute statistics.

## Installing

```
brew tap ch1aki/homebrew-gost
brew install ch1aki/gost/gost
```

or

```
go install github.com/ch1aki/gost@latest
```

## Usage

```
Usage:
  gost [OPTIONS] [FILE...]

Application Options:
  -v, --version     show version
  -f, --format=     fmt format for float columns (default: fixed-point decimal)
  -N, --count       sample size
      --min         minimum
      --max         maximum
      --sum         sum of elements of the sample
  -m, --mean        mean
      --sd          standard deviation
      --variance    variance
      --population  population variance and standard deviation
      --markdown    markdown table format

Help Options:
  -h, --help        Show this help message
```

Numbers come from files, or from standard input when there are no file
arguments or an argument is `-`. Several files are one sample. Each line may
hold several whitespace-separated numbers. Blank lines are skipped.

A bad token, `NaN`, or `Inf` is an error on stderr, `<file>:<line>: <reason>`,
and gost exits 1 without a table. A line may contain many numbers. Only a line
longer than 8MiB is `<file>:<line>: line too long`. Standard input is named
`stdin`. A missing file is `<name>: no such file or directory`. A directory is
`<name>: is a directory`. If nothing valid was read, gost exits 1 with
`<name>: no values`, listing every input when there are several.

With no column flags, gost prints N, MIN, MAX, SUM, MEAN, and STDDEV.
`--variance` adds VARIANCE. STDDEV and VARIANCE are the sample statistics
(divide by n−1). For one value those two cells are `NA`, because the sample
statistic is undefined. `--population` divides by n instead; one value is then
0. N is always an integer. Floats are fixed-point decimals unless `-f` is set.
`-f` is then a fmt verb, such as `%.2f` or `%g`.

```
$ seq 1 100 | gost
N  	MIN	MAX	SUM 	MEAN	STDDEV            
100	1  	100	5050	50.5	29.011491975882016	

$ seq 1 100 | gost --population
N  	MIN	MAX	SUM 	MEAN	STDDEV            
100	1  	100	5050	50.5	28.86607004772212	

$ printf '5\n' | gost
N	MIN	MAX	SUM	MEAN	STDDEV 
1	5  	5  	5  	5   	NA    	

$ printf '1 2 3\n4\n' | gost -f '%.2f' -N -m
N	MEAN 
4	2.50	
```

## Authors

ch1aki

## License

Licensed under the MIT License. See [LICENSE](LICENSE).
