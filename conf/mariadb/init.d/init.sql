GRANT ALL PRIVILEGES ON `money`.* to 'money'@'%' IDENTIFIED BY 'money';

CREATE TABLE `accounts` (
  `id` VARCHAR(40) DEFAULT (uuid()) NOT NULL,
  `name` VARCHAR(100) NOT NULL,
  `type` VARCHAR(20) NOT NULL,
  UNIQUE KEY `account_id` (`id`),
  UNIQUE KEY `account_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

CREATE TABLE `account_expected_balances` (
  `account_id` VARCHAR(40) NOT NULL,
  `timestamp` DATETIME NOT NULL,
  `expected_balance` FLOAT(2) NOT NULL,
  `import_id` VARCHAR(40) DEFAULT NULL, -- identifies all records imported at the same time with a UUID (could be linked to other table describing who did it, file name, time when imported, etc...)
  UNIQUE KEY `account_expected_balances` (`account_id`,`timestamp`),
  FOREIGN KEY (`account_id`) REFERENCES accounts(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

CREATE TABLE `period_transactions` (
  `id` VARCHAR(40) DEFAULT (uuid()) NOT NULL,
  `timestamp` DATETIME NOT NULL,
  `dt_account_id` VARCHAR(40) NOT NULL,
  `ct_account_id` VARCHAR(40) NOT NULL,
  `amount` FLOAT(2) NOT NULL,
  `details` VARCHAR(200) DEFAULT NULL,
  `summary` VARCHAR(200) DEFAULT NULL, -- assembled for search on ts, amount, debit and credit account names and details
  `import_id` VARCHAR(40) DEFAULT NULL, -- identifies all records imported at the same time with a UUID (could be linked to other table describing who did it, file name, time when imported, etc...)
  FOREIGN KEY (`dt_account_id`) REFERENCES accounts(`id`),
  FOREIGN KEY (`ct_account_id`) REFERENCES accounts(`id`),
  UNIQUE KEY `transaction_id` (`id`),
  UNIQUE KEY `transaction_search` (`summary`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;
