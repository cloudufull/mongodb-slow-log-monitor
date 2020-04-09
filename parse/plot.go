package parse

import (
       "gonum.org/v1/plot"
       "math"
       "fmt"
)

var default_SuggestedTicks float64 = float64(3)


func SetTicks(n int64){
     default_SuggestedTicks= float64(n)
}

func GetTicks() float64{
     return default_SuggestedTicks 
}


type MyTicks struct{}

var _ plot.Ticker = MyTicks{}

// Ticks returns Ticks in the specified range.
func (MyTicks) Ticks(min, max float64) (ticks []plot.Tick) {
	SuggestedTicks := GetTicks() 
	tens := math.Pow10(int(math.Floor(math.Log10(max - min))))
	n := (max - min) / tens
	for n < SuggestedTicks {
		tens /= 10
		n = (max - min) / tens
	}
	majorMult := int(n / SuggestedTicks)
	switch majorMult {
	case 7:
		majorMult = 6
	case 9:
		majorMult = 8
	}
	majorDelta := float64(majorMult) * tens
	val := math.Floor(min / majorDelta) * majorDelta
	for val <= max {
		if val > min && val < max {
			ticks = append(ticks, plot.Tick{Value: val, Label: fmt.Sprintf("%g", float32(val)) })
		}
		val += majorDelta
	}

	minorDelta := majorDelta/2
	switch majorMult {
	case 3, 6:
		minorDelta = majorDelta/3
	case 5:
		minorDelta = majorDelta/5
	}

	val = math.Floor(min / minorDelta) * minorDelta
	for val <= max {
		found := false
		for _, t := range ticks {
			if t.Value == val {
				found = true
			}
		}
		if val > min && val < max && !found {
			ticks = append(ticks, plot.Tick{Value: val})
		}
		val += minorDelta
	}
        ticks[0]=plot.Tick{Value: min, Label: fmt.Sprintf("%g", float32(min))}
        if SuggestedTicks>3 && max>min {
           max_lidx:=0
           for i,v:=range ticks{
               if v.Label!=""{
                  max_lidx=i
               }
           }
           ticks[max_lidx]=plot.Tick{Value: max, Label: fmt.Sprintf("%g", float32(max))}
        }else if SuggestedTicks==2{
           ticks[len(ticks)-1]=plot.Tick{Value: max, Label: fmt.Sprintf("%g", float32(max))}
        }
        //fmt.Println("================================================================")
        //for _,k:=range ticks{
        //    fmt.Printf("%#v\n",k)
        //}
	return
}
