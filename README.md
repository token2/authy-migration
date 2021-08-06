# Authy Migration Toolset


Forked from [![GoDoc](https://godoc.org/github.com/alexzorin/authy?status.svg)](https://godoc.org/github.com/alexzorin/authy)

This is a Go library that allows you to access your [Authy](https://authy.com) TOTP tokens.

It was created to facilitate migrating from Authy to Token2 hardware tokens (including Molto2 multi-profile TOTP hardware token) or other TOTP Apps.


Please note that this tool only migrates non-Authy-hosted accounts (the ones that are generating 7-digit OTP with 10/20 seconds interval). The tool is intended to migrate "standard" TOTP profiles : 6 or 8 digits, 30 seconds (Authy app supports only 30 seconds TOTP profiles in addition to its native accounts).

## Applications

### authy-export
This program will enrol itself as an additional device on your Authy account and export all of your TOTP tokens in [Key URI Format](https://github.com/google/google-authenticator/wiki/Key-Uri-Format).

**Installation**

Pre-built binaries are available from the [page](https://www.token2.swiss/site/page/how-to-transfer-totp-profiles-from-authy-to-a-token2-hardware-token).

Alternatively, it can be compiled from source, which requires [Go 1.12 or newer](https://golang.org/doc/install):

```shell
go get github.com/token2/authy-migration/cmd/authy-export
```

**To use it:**

1. Run `authy-export`
2. The program will prompt you for your phone number country code (e.g. 1 for United States) and your phone number. This is the number that you used to register your Authy account originally.
3. If the program identifies an existing Authy account, it will send a device registration request using the `push` method. This will send a push notification to your existing Authy apps (be it on Android, iOS, Desktop or Chrome), and you will need to respond that from your other app(s).
4. If the device registration is successful, the program will save its authentication credential (a random value) to `$HOME/authy-go.json` for further uses. **Make sure to delete this file and de-register the device after you're finished.**
5. If the program is able to fetch your TOTP encrypted database, it will prompt you for your Authy backup password. This is required to decrypt the TOTP secrets for the next step. 
6. The program will dump all of your TOTP tokens in a file in the same folder: a .txt file which can be used for importing to Molto2 directly, or HTML file with QR codes, which you can use to import to other applications.

## Third-party modules
If you wish to compile this from the source code, make sure you add the following modules to your Go enviroment

``` 
golang.org/x/crypto/ssh/terminal
github.com/skip2/go-qrcode
```


## LICENSE

See [LICENSE](LICENSE)

## Trademark Legal Notice

All product names, logos, and brands are property of their respective owners. All company, product and service names used in this website are for identification purposes only. Use of these names, logos, and brands does not imply endorsement
