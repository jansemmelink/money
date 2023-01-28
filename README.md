# Money #

# Progress
* Created db schema with accounts and transactions and expected account balances
* Wrote import utility to:
    * imports StdBank CSV files (debit or credit card) from console into database,
    * detects duplicates and undo the import, allowing you to fix the CSV then import again
    * it also imports expected open/closing balances
    * you can also delete overlapping parts of the statement, but then open/close balance will not work - need to consider skipping the sum check then... but open/close balances will still remove, or just ignore overlapping records after import and keep the rest... but be careful not to miss duplicate bank fee entries or duplicate payments on the same day etc... which I saw exists in CSV files even when I used other references when making duplicate payments from the banking app!
    * After import, all transactions (expect bank fees) are against unknown expenses/income account. One need to find and set all of them against actual accounts to one can then see where the money went or originated from

# Next
* Utility to update transactions with unknown income or expense to some other account
    * start with filter on which transaction to target (e.g. most recent)


    * Ideally show the transaction and select the other account from a list, or chose to create a new account, or choose to skip the tx for now.
    * Create accounts with sub items, e.g. "hardware|paint house" then report on hardware or specifically painting the house
    * See need to select account and sub account or use parent or add sub or ...
    * Indicate expense/income - but should know that from the type of transaction (although it could be payment of an asset too I suppose)
    * Show nr of transactions still using unknown
* Generate reports totals and per account, summary or detailed, ...
    * list dates for which transactions are available or missing
    * identify transfers between accounts (e.g. pay credit card from cheque acc)
    * balance over all bank accounts over time + sum value

* Generate graphs over time - use spreadsheet for that or ...?
* Consider doing an SQL transaction around import then commit or rollback on duplicates
* Test import with partial overlapping statements
* Import as many as possible historic statements (see what I have in dropbox...)
* Add cash transactions manually (todo mobile app)
* See if can automate regular imports from banking email or FTP statements
* Make this permanent in some hosted env with secure access
* Make it multi-tenanted with web and mobile app
