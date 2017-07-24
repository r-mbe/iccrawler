package common

type Supplier struct {
	suppliers map[string]int
}

/*
    if _, ok := suppliers["aii"]; ok {
        fmt.Println("yes key exist!", ok)
    }

    for key, value := range suppliers {
		fmt.Printf("%s : %d\n", key, value)
	}
*/

func NewSP()( *Supplier)  {
	sp := new(Supplier)
	sp.Init()
	return &sp
}

func (sp *Supplier) Init() {
	this.suppliers = map[string]int{
		"advancedmp":  29,
		"aii":         6,
		"aipco":       27,
		"avnet":       14,
		"bristol":     16,
		"chip1stop":   1,
		"digikey":     2,
		"element14":   26,
		"element14cn": 34,
		"future":      3,
		"hdi":         3282,
		"ickey":       30,
		"microchip":   10,
		"ps":          11,
		"rutronik":    12,
		"vicor":       13,
		"wpi":         4,
		"rochester":   33,
		"master":      7,
	}
}

func (sp *Supplier) hasSupplier(key) (bool, error) {

	if _, ok := sp.suppliers[key]; ok {
		//yes it has
		return true, nil
	}

	return false, !nil
}
