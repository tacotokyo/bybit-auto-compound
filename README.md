# bybit-auto-compound

This CLI maintain Inverse Perpetual short position into leverage x1 in order to make delta-neutral. Only works when funding rate is positive. This is useful when you are receiving funding and re-invest it.

## Usage

1. Download binary from [here](https://github.com/tacotokyo/bybit-auto-compound/releases)

```
mkdir bybit && cd bybit
wget https://github.com/tacotokyo/bybit-auto-compound/releases/download/v0.4/bybit-auto-compound-linux
chmod +x bybit-auto-compound-linux
```

2. Copy `.env.example` file into `.env` and update your key and secret

```
wget -O .env https://raw.githubusercontent.com/tacotokyo/bybit-auto-compound/main/.env.example
```

3. Setup cron
```
*/10 * * * * /root/bybit/bybit-auto-compound-linux -coin xrp,eos -envfile /root/bybit/.env 2>/root/bybit/bybit-auto-compound.log
```
