# bettercsv

Bettercsv is an alternative to the native Go csv. It provides several additional features:
- A method for retrieving the headers.
- Allows reading to maps with the headers as keys and the fields as values.
- Allows gracefully handling errors to continue reading on error.

```
// New Attributes:
  SkipLineOnErr  bool // Skips line when error occurs, allowing reader to continue

// New Methods:
  func (r *Reader) Headers() (headers []string, err error)
  func (r *Reader) ReadToMap() (recordMap map[string]string, err error)
  func (r *Reader) ReadAllToMaps() (records []map[string]string, err error)
  func (r *Reader) ReadAllWithErrors() (records [][]string, errs []error)
  func (r *Reader) ReadAllToMapsWithErrors() (records []map[string]string, errs []error)
```

### Headers
text.csv

```
first,last,email
John,Doe,john@doe.com
Jane,Doe,jane@doe.com
```

Initialize our reader:

```
csvfile, err := os.Open("text.csv")
reader := bettercsv.NewReader(csvfile)
```

Calling `reader.Headers()` will return `[first last email]`. _Note: Calling `.Headers()` will advance the reader to the second line._

### ReadToMap(s)
Calling `reader.ReadToMap()` (after calling `.Headers()`) will return:

```
[first:John last:Doe email:john@doe.com]
```

You can call `reader.ReadAllToMaps()` to return a slice of `map[string]string`.

## Error Handling

When reading line by line using `reader.Read()`, if an error occurs, `csv` will continue reading from the error and you will receive a cascade of errors. For example:

### Example

text.csv


```
first,last,email
John,Doe,john@doe.com
Jane,Doe",jane@doe.com
June,Doe,june@doe.com
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
[June Doe june@doe.com]
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

### bettercsv
With `reader.SkipLineOnErr = true`.


```
[first last email]
[John Doe john@doe.com]
line 3, column 22: bare " in non-quoted-field
[]
[June Doe june@doe.com]
line 5, column 23: bare " in non-quoted-field
[]
line 6, column 0: wrong number of fields in line
[Jim Doe jim doe.com]
[Joan Doe joan@doe.com]
line 8, column 23: extraneous " in field
[]
[Jill Doe jill@doe.com]
```

Notice the line numbers for bettercsv point to their correct lines and Jill Doe was removed completely from the standard csv library.

### ReadWithErrors

If you prefer to use the `reader.ReadAll()` method but still want to skip errors, you can use `reader.ReadAllWithErrors()` to receive both a slice of slices of records and a slice of all the errors.


```
values:
[[John Doe john@doe.com]
 [June Doe june@doe.com]
 [Joan Doe joan@doe.com]
 [Jill Doe jill@doe.com]]

errors:
[line 3, column 22: bare \" in non-quoted-field
line 5, column 23: bare " in non-quoted-field
line 6, column 0: wrong number of fields in line
line 8, column 23: extraneous " in field]
```

You can also combine errors and maps with `reader.ReadAllToMapsWithErrors()`.

