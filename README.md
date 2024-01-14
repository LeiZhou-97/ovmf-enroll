# A tool for ovmf volumes

This is a tiny tool for EDK2 firmware image, which supports decoding and printing the variable list of Variable Stores.

It also supports to add new variable into Vriable Stores.

# How-to use

## Build executable binary

```bash
go build -o ovmfctl ./cmd/main.go
```

## Enroll vairbale

```bash
./ovmfctl  -f <path/to/input/OVMF.fd> -o <path/to/output/OVMF.fd> -n <varibale_name> -g <variable_guid>  -a <variable_attributes> -d <varibale_data_file>
```


## Todo List

- Support delete variable
- Support secure boot
