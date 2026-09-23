<div align="center">
  
<img width="96" alt="preview" src="https://imgur.com/5ldlRqF.png" />

# Mac Spoofer

**Wi-Fi MAC address spoofer that auto-detects wireless adapters**

https://github.com/user-attachments/assets/cf69c331-a8ab-4971-b9de-2c5861e9fb29

</div>

 ## Options

1. **Spoof MAC address**
2. **Revert back to your burned-in hardware MAC**
3. **Check MAC address**
4. **Exit**

## Needs

- Windows + Wi-Fi adapter
- admin elevation

# Download
you can find the download from the [releases](https://github.com/yourpovv/Mac-Spoofer/releases/tag/V1.0.0)

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
