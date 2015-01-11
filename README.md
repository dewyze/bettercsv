# skippycsv

Skippycsv is an alternative to the native Go csv. It allows for skipping the rest of the line on an error.

### Problem

When reading line by line using reader.Read(), if an error occurs `csv` will continue reading from the error and you will receive a cascade of errors. For example:

## Example

test.csv


```
first,last,email
Jane,Doe",jane@doe.com
Joan,Doe,joan@doe.com
Jeff,D"oe",jeff@doe.com
Jim,Doe,jim,doe.com
Joan,Doe,joan@doe.com
Jack,"Do"e",jack@doe.com
Jill,Doe,jill@doe.com
```

test.go

```go
...
for {
  record, err := reader.Read()
  if err != nil {
    if err == io.EOF {
      break
    }

    fmt.Println(err)
  }
  fmt.Println(record)
}
...
```

Looping over this with `reader.Read()` and printing each line will result in:

### encoding/csv

```
[first last email]
[John Doe john@doe.com]
line 3, column 8: bare " in non-quoted-field
[]
line 4, column 0: wrong number of fields in line
[ jane@doe.com]
[Joan Doe joan@doe.com]
line 6, column 6: bare " in non-quoted-field
[]
line 7, column 2: bare " in non-quoted-field
[]
line 8, column 0: wrong number of fields in line
[ jeff@doe.com]
line 9, column 0: wrong number of fields in line
[Jim Doe jim doe.com]
[Joan Doe joan@doe.com]
line 11, column 8: extraneous " in field
[]
line 14, column 0: extraneous " in field
[]
```

### skippycsv

```
[first last email]
[John Doe john@doe.com]
line 3, column 22: bare " in non-quoted-field
[]
[Joan Doe joan@doe.com]
line 5, column 23: bare " in non-quoted-field
[]
line 6, column 0: wrong number of fields in line
[Jim Doe jim doe.com]
[Joan Doe joan@doe.com]
line 8, column 23: extraneous " in field
[]
[Jill Doe jill@doe.com]
```

Notice the line numbers for skippcsv point to their correct lines and Jill Doe was removed completely from the standard csv library.

To use skippycsv simply set the following to your reader object:

```
reader := skippycsv.NewReader(csvfile)
reader.SkipLineOnErr = true
```
