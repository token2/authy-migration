# Authy Migration Toolset


* Update 03/07/2024 : It appears that Authy is blocking some of the requests, which unfortunately impairs the functionality of our tool. We are investigating this, but until further notice, this toolset is not usable."


Forked from [![GoDoc](https://godoc.org/github.com/alexzorin/authy?status.svg)](https://godoc.org/github.com/alexzorin/authy)

This is a Go library that allows you to access your [Authy](https://authy.com) TOTP tokens.

It was created to facilitate migrating from Authy to Token2 hardware tokens (including Molto2 multi-profile TOTP hardware token) or other TOTP Apps. Additionally, it supports exporting profiles to [WinAuth](https://winauth.github.io/winauth/index.html), a desktop TOTP app, and as Raivo JSON export which can be imported by, say, [2FAS](https://2fas.com) an Open Source app for iOS and Android.


Please note that this tool only migrates non-Authy-hosted accounts (the ones that are generating 7-digit OTP with 10/20 seconds intervals). The tool is intended to migrate "standard" TOTP profiles: 6 or 8 digits, 30 seconds (Authy app supports only 30 seconds TOTP profiles in addition to its native accounts).

## Applications

### authy-export
This program will enrol itself as an additional device on your Authy account and export all of your TOTP tokens in [Key URI Format](https://github.com/google/google-authenticator/wiki/Key-Uri-Format).

**Installation**

Pre-built binaries are available from the [page](https://www.token2.swiss/site/page/how-to-transfer-totp-profiles-from-authy-to-a-token2-hardware-token).

Alternatively, it can be compiled from source, which requires [Go 1.12 or newer](https://golang.org/doc/install):

Follow the steps below to run the script from source code:

Install [Go 1.12 or newer](https://golang.org/doc/install)

Download the latest version of the [toolset](https://github.com/token2/authy-migration/archive/refs/heads/master.zip):

```shell
git clone https://github.com/token2/authy-migration.git
cd authy-migration
```

Install the qrcode module (other required modules will be installed automatically on the first run)

```shell
go get github.com/skip2/go-qrcode
```

Unzip to a folder, launch command line and change folder to {path}/cmd/authy-export and launch the command

```shell

go run authy-export.go

```

**To use it:**

1. Launch `authy-export.go`
2. The program will prompt you for your phone number country code (e.g. 1 for United States) and your phone number. This is the number that you used to register your Authy account originally.
3. If the program identifies an existing Authy account, it will send a device registration request using the `push` method. This will send a push notification to your existing Authy apps (be it on Android, iOS, Desktop or Chrome), and you will need to respond that from your other app(s).
4. If the device registration is successful, the program will save its authentication credential (a random value) to `$HOME/authy-go.json` for further uses. **Make sure to delete this file and de-register the device after you're finished.**
5. If the program is able to fetch your TOTP encrypted database, it will prompt you for your Authy backup password. This is required to decrypt the TOTP secrets for the next step. 
6. The program will dump all of your TOTP tokens in a file in the same folder: a **.txt** file which can be used for importing to Molto2 directly, or an **.html** file with QR codes, which you can use to import to other applications. If you plan to import the TOTP profiles exported from Authy to WinAuth, the filename should end with **.wa.txt** extension. For Raivo export the program will create a **.raivos** file which you rename to **.json** after possibly editing it to adjust account names.

## Third-party modules
The module below is used for generating QR codes
 
[github.com/skip2/go-qrcode](https://github.com/skip2/go-qrcode)
 


## LICENSE

See [LICENSE](LICENSE)

## Trademark Legal Notice

All product names, logos, and brands are property of their respective owners. All company, product and service names used in this website are for identification purposes only. Use of these names, logos, and brands does not imply endorsement
