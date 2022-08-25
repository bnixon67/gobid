CREATE OR REPLACE VIEW current_bids AS
SELECT a.id, a.created, a.bidder, a.amount
FROM bids a
INNER JOIN (
  SELECT id, MAX(amount) amount
  FROM bids
  GROUP BY id
) b ON a.id = b.id AND a.amount = b.amount;
