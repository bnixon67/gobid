DELIMITER //

-- placeBid will try and place a bid for an item. 
CREATE OR REPLACE PROCEDURE placeBid(
  bidId int(11),
  newAmount decimal(13,2),
  newBidder varchar(30)
)
MODIFIES SQL DATA
BEGIN
  DECLARE bidPlaced boolean DEFAULT false;
  DECLARE minAmount decimal(13,2) DEFAULT NULL;
  DECLARE openingBid decimal(13,2) DEFAULT 0;
  DECLARE minBidIncr decimal(13,2) DEFAULT 0;
  DECLARE curBidder varchar(30) DEFAULT "";
  DECLARE curAmount decimal(13,2) DEFAULT 0;
  DECLARE message varchar(30);

  START TRANSACTION;

  -- ensure item exists
  SELECT COUNT(*) INTO @cnt FROM items WHERE id = bidId;
  IF @cnt = 0 THEN
    SET message = 'No such item';
  ELSEIF @cnt > 1 THEN
    SET message = 'Multiple rows';
  ELSE
    -- get current bid information
    SELECT items.openingBid, items.minBidIncr,
           current_bids.bidder, current_bids.amount
    INTO openingBid, minBidIncr, curBidder, curAmount
    FROM items LEFT OUTER JOIN current_bids ON items.id = current_bids.id
    WHERE items.id = bidId
    FOR UPDATE; -- lock tables within transaction

    IF openingBid = 0 THEN
      SET message = 'Display only item';
    ELSE
      SET minAmount = IF(ISNULL(curAmount),
                         openingBid,
                         curAmount+minBidIncr);

      IF newAmount < minAmount THEN
        SET message = 'Bid too low';
      ELSE
        INSERT INTO bids(id, bidder, amount)
        VALUES(bidId, newBidder, newAmount);

        SET @rows = row_count();
        IF @rows != 1 THEN
          SET message = 'No rows inserted';
        ELSE 
	  SET bidPlaced = true;
          SET message = 'Bid placed';
        END IF;
      END IF;
    END IF;
  END IF;

  SELECT bidPlaced, bidId, message,
         IFNULL(curBidder,"") AS priorBidder,
         IFNULL(minAmount,0) AS minAmount,
	 newBidder, newAmount;

  COMMIT;

END //

DELIMITER ;
