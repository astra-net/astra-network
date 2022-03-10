# Blacklist info

The black list is a newline delimited file of wallet addresses. It can also support comments with the `#` character.

## Default Location

By default, the astra binary looks for the file `./.astra/blaklist.txt`.

## Example File

```
0xef1c0d949efbd8baed211fad28a08c5cf96e989d
0x3b00aab412891853f2cc91a6957a66088f22437b  # This is a comment
0xafdd04699c52635b059a4b914f61253b0b0093ad

```

## Details

Each transaction added to the tx-pool has its `to` and `from` address checked against this blacklist.
If there is a hit, the transaction is considered invalid and is dropped from the tx-pool.
