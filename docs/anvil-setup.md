# Anvil Setup GreenTrust Passport

Dokumen ini dipakai untuk setup ulang blockchain lokal dari nol ketika Anvil mati, database di-reset, atau contract address berubah.

## 1. Cek Foundry

Pastikan `anvil`, `forge`, dan `cast` tersedia.

```bash
anvil --version
forge --version
cast --version
```

Kalau belum ada, install Foundry dulu:

```bash
curl -L https://foundry.paradigm.xyz | bash
foundryup
```

Tutup terminal lalu buka ulang kalau command Foundry belum kebaca.

## 2. Jalankan Anvil

Dari terminal baru:

```bash
anvil
```

Default Anvil akan berjalan di:

```text
http://127.0.0.1:8545
```

Chain ID default:

```text
31337
```

Pakai private key account pertama Anvil untuk deploy dan backend:

```text
0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
```

Jangan tutup terminal Anvil selama backend dipakai. Kalau terminal ini ditutup, state blockchain lokal hilang kecuali Anvil dijalankan dengan state persistence.

## 3. Deploy Contract

Dari root project:

```bash
cd contracts
forge build
forge script script/DeployGreenPassport.s.sol:DeployGreenPassportRegistry \
  --rpc-url http://127.0.0.1:8545 \
  --private-key 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80 \
  --broadcast
```

Setelah deploy, cari output seperti:

```text
GreenPassportRegistry deployed at: 0x...
```

Atau lihat bagian `Deployed to`.

Pada Anvil yang fresh, kalau belum ada transaksi lain, address biasanya:

```text
0x5FbDB2315678afecb367f032d93F642f64180aa3
```

Tapi tetap gunakan address terbaru dari output deploy.

## 4. Update `.env` Backend

Di root project, update nilai berikut:

```env
BLOCKCHAIN_RPC_URL=http://127.0.0.1:8545
BLOCKCHAIN_PRIVATE_KEY=0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
BLOCKCHAIN_CHAIN_ID=31337
BLOCKCHAIN_NETWORK_NAME=Local Anvil
GREEN_PASSPORT_CONTRACT_ADDRESS=ISI_ADDRESS_HASIL_DEPLOY
BLOCKCHAIN_TX_TIMEOUT_SECONDS=90
```

Contoh kalau deploy address-nya default:

```env
GREEN_PASSPORT_CONTRACT_ADDRESS=0x5FbDB2315678afecb367f032d93F642f64180aa3
```

Penting: contract owner adalah wallet yang deploy contract. Karena deploy pakai private key account pertama Anvil, backend juga harus pakai private key yang sama supaya `issuePassport` tidak gagal `NotOwner`.

## 5. Restart Backend

Restart backend supaya env baru kebaca.

```bash
go run cmd/app/main.go
```

Kalau kamu pakai command lain untuk menjalankan server, cukup stop lalu start ulang.

## 6. Smoke Test Contract

Cek owner contract:

```bash
cast call \
  $GREEN_PASSPORT_CONTRACT_ADDRESS \
  "owner()(address)" \
  --rpc-url http://127.0.0.1:8545
```

Kalau command di atas dijalankan tanpa env shell, isi address manual:

```bash
cast call \
  0x5FbDB2315678afecb367f032d93F642f64180aa3 \
  "owner()(address)" \
  --rpc-url http://127.0.0.1:8545
```

Expected owner untuk account pertama Anvil:

```text
0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266
```

## 7. Flow Setelah Reset Anvil

Kalau Anvil mati dan kamu nyalakan ulang tanpa state persistence:

1. Jalankan `anvil`.
2. Deploy ulang contract dengan `forge script ... --broadcast`.
3. Copy contract address baru ke `.env`.
4. Restart backend.
5. Issue Green Passport lagi.

Data passport lama di database masih menyimpan tx hash dan block lama, tapi blockchain lokal yang baru tidak mengenal tx tersebut. Untuk demo paling aman, setelah reset Anvil gunakan data baru atau issue ulang passport setelah contract baru siap.

## 8. Opsional: Simpan State Anvil

Kalau tidak mau deploy ulang setiap Anvil restart, jalankan Anvil dengan state file:

```bash
anvil --dump-state anvil-state.json --load-state anvil-state.json
```

Catatan: file state bisa membesar dan sebaiknya tidak perlu dicommit kecuali memang ingin menyimpan snapshot demo.

## Troubleshooting

### `blockchain client disabled: blockchain env is incomplete`

Pastikan `.env` punya:

```env
BLOCKCHAIN_RPC_URL
BLOCKCHAIN_PRIVATE_KEY
BLOCKCHAIN_CHAIN_ID
GREEN_PASSPORT_CONTRACT_ADDRESS
```

Lalu restart backend.

### `connection refused`

Anvil belum berjalan atau bukan di port `8545`.

```bash
anvil
```

### `NotOwner`

Private key backend beda dengan wallet yang deploy contract. Deploy ulang contract dengan private key yang sama seperti `BLOCKCHAIN_PRIVATE_KEY`, atau ganti backend private key ke owner contract.

### `PassportAlreadyIssued`

Passport ID yang sama sudah pernah di-issue di contract state saat ini. Gunakan passport baru atau reset state Anvil dan deploy ulang.

### `EmptyDocumentHashes`

Backend mencoba issue passport tanpa document hash. Pastikan evidence required sudah ada dan statusnya `reviewed` atau `on_chain`.
