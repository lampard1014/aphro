package Biz

import (
	"github.com/lampard1014/aphro/Biz/room/pb"
	"time"
	"strings"
	"strconv"
	"math"
)

type IntervalUnit uint32

const (
	IntervalUnitInfinite IntervalUnit = iota
	IntervalUnitSecond
	IntervalUnitMinute
	IntervalUnitHour
	IntervalUnitDay
	IntervalUnitWeek
	IntervalUnitMonth
	IntervalUnitYear
)

const (
	TransactionCalculatorErrorMsgIntervalUnitPaserError = "[TransactionCalculatorError] Interval Unit Paser Error"
	TransactionCalculatorErrorMsgEndTimeGoingByError = "[TransactionCalculatorError] EndTime is going  by Aka Expired.."
	TransactionCalculatorErrorMsgEntityError = "[TransactionCalculatorError] Entity Error.."
	TransactionCalculatorErrorCalcaulateTimeError = "[TransactionCalculatorError] Mismatch Time When Calculate ..."

)

type  TransactionCalculatorError struct {
	Message string
}

func (this *TransactionCalculatorError) Error()string {
	return this.Message
}

func (static TransactionCalculator) newTransactionCalculatorError (message string)error{
	return &TransactionCalculatorError{message}
}

type TCRule  struct {
	fee float32
	start string
	end string
	interval uint32
	intervalUnit uint32
	merchant uint32
	roomID uint32
	transactionID uint32
	flag uint32
	rcrID uint32

}


type TransactionCalculator struct {
	//sema chan struct{}

}

//reformer to standard rule model
func (static TransactionCalculator)ReformerRuleByRCRCreatePB(in *Aphro_Room_pb.RCRCreateRequest) (*TCRule) {
	return &TCRule{
		in.Fee,
		in.Start,
		in.End,
		in.Interval,
		in.IntervalUnit,
		in.MerchantID,
		in.RoomID,
		0,
		in.Flag,
		0,
	}
}

func (static TransactionCalculator)fetchTimerBy(rule *TCRule)(t *time.Timer,err error) {
	//var uint time.Duration
	uint, err := static.fetchTimeDuration(IntervalUnit(rule.interval))
	if err != nil {
		return
	}
	//create Timer
	t = func(rule *TCRule) *time.Timer {
		return time.NewTimer(time.Duration(rule.interval) * uint)
	}(rule)
	return
}

func (static TransactionCalculator)fetchTimeDuration(bitInterval IntervalUnit)(d time.Duration,e error) {

	switch bitInterval{
		case IntervalUnitInfinite:
			d = 0
		case IntervalUnitSecond:
			d = time.Second
		case IntervalUnitMinute:
			d = time.Minute
		case IntervalUnitHour:
			d = time.Hour
		case IntervalUnitDay:
			d = time.Hour
		case IntervalUnitWeek:
			d = time.Hour
		case IntervalUnitMonth:
			d = time.Hour
		case IntervalUnitYear:
			d = time.Hour
		default:
			d = 0
			e = static.newTransactionCalculatorError(TransactionCalculatorErrorMsgIntervalUnitPaserError)
	}
	return
}

func (static TransactionCalculator)fetchSema(cap int)(sema chan struct{}) {
	return make (chan struct{}, cap)
}

//func (static TransactionCalculator)ScheduleRules(rules map[*TCRule]func(<-chan time.Time)) (err error) {
//}

//schedule timer
func (static TransactionCalculator)ScheduleRules(rules map[*TCRule]func(<-chan time.Time)) (err error){

	//sema := static.fetchSema(len(rules))

	for r,f := range rules {
		var t *time.Timer
		t,err =static.fetchTimerBy(r)
		if err == nil {
			go func (){
				//sema <- struct{}{}
				f(t.C)
				//<-sema
			}()
		}

	}
	return err
}

//计算费用，从begin 到当前时间的完整费用值
func (static TransactionCalculator)CalculateFee(rule *TCRule,begin time.Time ,c <-chan time.Time) (currentFee float64, err error) {


	if rule == nil {
		err = static.newTransactionCalculatorError(TransactionCalculatorErrorMsgEntityError)
	} else {
		var uint time.Duration
		uint, err = static.fetchTimeDuration(IntervalUnit(rule.interval))
		if err != nil {
			return
		}
		interval := float64(rule.interval)
		stHour, stMin, stSecond := static.parseLimitTime(rule.start)
		etHour, etMin, etSecond := static.parseLimitTime(rule.end)
		//现在离开始的时间差
		t := <-c
		timeDuration := t.Sub(begin)

		//目前只考虑时，分，秒
		switch uint {
			case time.Hour:
				nh := t.Hour()
				if nh < stHour || ((etHour >=24 && nh > etHour - 24) || nh > etHour) {
					//还没开始 或者 已经结束
					err = static.newTransactionCalculatorError(TransactionCalculatorErrorCalcaulateTimeError)
				} else {
					//在范围之内
					calculateInterval := static.ceilTimeInterval(timeDuration.Hours(),interval)
					currentFee = calculateInterval * float64(rule.fee)
				}
			case time.Second:
				ns := t.Second()
				if ns < stSecond || ((etSecond >=60 && ns > etSecond - 60) || ns > etSecond) {
					//还没开始 或者 已经结束
					err = static.newTransactionCalculatorError(TransactionCalculatorErrorCalcaulateTimeError)
				} else {
					//在范围之内
					calculateInterval := static.ceilTimeInterval(timeDuration.Seconds(),interval)
					currentFee = calculateInterval * float64(rule.fee)
				}
			case time.Minute:
				ns := t.Minute()
				if ns < stMin || ((etMin >=60 && ns > etMin - 60) || ns > etMin) {
					//还没开始 或者 已经结束
					err = static.newTransactionCalculatorError(TransactionCalculatorErrorCalcaulateTimeError)
				} else {
					//在范围之内
					calculateInterval := static.ceilTimeInterval(timeDuration.Minutes(),interval)
					currentFee = calculateInterval * float64(rule.fee)
				}
			case 0:
				//无间隔,包场
				currtentTimeValue := t.Hour() * 3600 + t.Minute() * 60 + t.Second()
				stTimeValue := stHour * 3600 + stMin * 60 + stSecond
				etTimeValue := etHour * 3600 + etMin * 60 + etSecond
				if currtentTimeValue < stTimeValue || currtentTimeValue > etTimeValue {
					//还没开始 或者 已经结束
					err = static.newTransactionCalculatorError(TransactionCalculatorErrorCalcaulateTimeError)
				} else {
					//在范围之内
					currentFee = float64(rule.fee)
				}
		}
	}

	return
}

func (static TransactionCalculator)parseLimitTime(lt string)(hour int,minute int, second int) {

	//var err error
	t:= strings.Split(lt,":")
	if len(t) == 3 {
		hour,_ = strconv.Atoi(t[0])
		minute,_ = strconv.Atoi(t[1])
		second,_ = strconv.Atoi(t[2])
	}
	return
}

func (static TransactionCalculator)ceilTimeInterval(t float64, uint float64)float64 {
	return (math.Ceil(t/uint))
}


type TransactionCalculatorErr struct  {
	code int
	message string
}

func (this *TransactionCalculatorErr) Error ()string{
	return this.message
}
