package NSUtil

// price type
const (
	Fix = iota
    Auction
    OnlyShow
)

// product type
const (
    Digit = iota
    Entity
)

// order state
const (
	None = iota			
	InFix	
    InAuction
    Unshipped               // product isn't shipped
    Dispatched              // product is shipped
    DispatchConfirmed       // buyer confirms transaction
	Completed 			    // transaction is successful
    ReturnInAgree           // waiting for agreeing from seller
    ReturnAgreed            // this return is agreed by seller
    ReturnDispatched        // product is returned by buyer
    ReturnConfirmed         // seller receives returned product
    ReturnCompleted         // return is completed
    Failed                  // transaction fails
)