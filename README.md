<div align="center">

# Mac Spoofer

**Wi-Fi MAC address spoofer that auto-detects wireless adapters**


</div>

## Options

`1` - spoof mac address
`2` - revert back to your burned-in hardware MAC
`3` - check mac address
`4` - exit

## Needs

- Windows + Wi-Fi adapter
- admin elevation

## Run

```bash
go run .
```

building it:

```bash
go build -o mac-spoofer.exe .
./mac-spoofer
```

## Known issues

- some Realtek driver versions ignore the spoof so update the driver and try again if the MAC doesn't change

## License

[MIT](LICENSE) © [YourPOV](https://github.com/yourpovv)
