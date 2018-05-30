package Biz

import (
	"github.com/lampard1014/aphro/Biz/room/pb"
	"time"
	"strings"
	"strconv"
	"math"
	"fmt"
	"github.com/lampard1014/aphro/PersistentStore/MySQL"
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

func (static TransactionCalculator)BatchReformerRuleByRCRCreatePB(in []*Aphro_Room_pb.RCRCreateRequest) ([]*TCRule) {
	rules := []*TCRule{}

	for _,v := range in {
		r := static.ReformerRuleByRCRCreatePB(v)
		rules = append(rules, r)
	}

	return rules
}


func (static TransactionCalculator)fetchTickerBy(rule *TCRule)(t *time.Ticker,err error) {
	//var uint time.Duration
	uint, err := static.fetchTimeDuration(IntervalUnit(rule.intervalUnit))
	if err != nil {
		return
	}
	t = time.NewTicker(time.Duration(rule.interval) * uint)
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


//创建timer
//timer的业务逻辑放到 goroutine 来处理
func (static TransactionCalculator)ScheduleRulesByRules(rules []*TCRule, begin time.Time) (currentFee float64,err error) {

	//biz logical here
	tickerFunc := func(ticker *time.Ticker, rule *TCRule)(fee float64,err error) {

		for t := range ticker.C {
			//step 1 get money
			fee ,err = static.CalculateFee(rule, begin, t)
			fmt.Println("log ",fee, err,time.Now())
			//step2 update database

			//step3 checkout need remote Ticker
			static.stopTicker(ticker,rule,t)
		}
		return
	}

	for _,r := range rules {
		err = static.ScheduleRulesByRuleAndFunc(r, tickerFunc)
		err = static.SnapRule(r)
		//assume calculate
		func (){
			f,_ := static.CalculateFee(r,begin,time.Now())
			currentFee += f
		}()
	}

	return
}

func (static TransactionCalculator)SnapRule(rule *TCRule) (err error) {

	var mysql *MySQL.APSMySQL
	mysql,err = MySQL.NewAPSMySQL(nil)
	if err == nil {
		m, ok := mysql.Connect().(*MySQL.APSMySQL)
		defer m.Close()

		if ok {

		}
	}
	return
}

//schedule timer
func (static TransactionCalculator)ScheduleRulesByRuleAndFunc(rule *TCRule,f func( *time.Ticker, *TCRule)(float64,error)) (err error){

	t, err := static.fetchTickerBy(rule)
	if err == nil {
		go f(t, rule)
	}
	return err
}

//remove tiker
func (static TransactionCalculator)stopTicker (ticker *time.Ticker, rule *TCRule, now time.Time) {


	var uint time.Duration
	uint, _ = static.fetchTimeDuration(IntervalUnit(rule.intervalUnit))
	interval := float64(rule.interval)

	etHour, etMin, etSecond := static.parseLimitTime(rule.end)

	tn := now.Hour() * 3600 + now.Minute() * 60 + now.Second()
	etTimeValue := etHour * 3600 + etMin * 60 + etSecond
	overflow := interval * float64(uint / time.Second)

	var beyondTime bool = false
	if etHour > 24 || etMin > 60 || etSecond > 60 {
		beyondTime = true
	}

	//超过一次overflow的时间算是过期
	if  (tn > etTimeValue + int(overflow)) || (beyondTime && (tn + 24 *3600) > etTimeValue + int(overflow))  {
		ticker.Stop()
	}
}

//计算费用，从begin 到当前时间的完整费用值
func (static TransactionCalculator)CalculateFee(rule *TCRule,begin time.Time ,now time.Time) (currentFee float64, err error) {
	if rule == nil {
		err = static.newTransactionCalculatorError(TransactionCalculatorErrorMsgEntityError)
	} else {
		var uint time.Duration
		uint, err = static.fetchTimeDuration(IntervalUnit(rule.intervalUnit))
		if err != nil {
			return
		}
		interval := float64(rule.interval)
		stHour, stMin, stSecond := static.parseLimitTime(rule.start)
		etHour, etMin, etSecond := static.parseLimitTime(rule.end)


		var beyondTime bool = false
		if etHour > 24 || etMin > 60 || etSecond > 60 {
			beyondTime = true
		}

		tn := now.Hour() * 3600 + now.Minute() * 60 + now.Second()

		stTimeValue := stHour * 3600 + stMin * 60 + stSecond
		etTimeValue := etHour * 3600 + etMin * 60 + etSecond

		overflow := interval * float64(uint / time.Second)

		var isBingo bool
		if (
			( tn >= stTimeValue && tn <= etTimeValue + int(overflow)) ||
			(beyondTime && (tn + 24 *3600) <= etTimeValue + int(overflow) && (tn + 24 *3600) >= stTimeValue) ) {
			isBingo = true
		}

		//默认今天 否则是前一天
		var tmp time.Time = now
		if tn < stTimeValue {
			p,_ := time.ParseDuration("-24h")
			tmp = now.Add(p)
		}
		startDateTime := time.Date(tmp.Year(),tmp.Month(),tmp.Day(),stHour,stMin,stSecond,0,tmp.Location())

		fmt.Println("xxxxx", startDateTime)
		mt := math.Max(float64(startDateTime.Unix()),float64(begin.Unix()))
		between := now.Unix() - int64(mt) //精确到秒

		if isBingo {
			calculateInterval := static.ceilTimeInterval(float64(between) *float64(time.Second) ,interval * float64(uint))

			if uint == 0 {
				currentFee = float64(rule.fee)
			} else {
				currentFee = calculateInterval * float64(rule.fee)
			}
		} else {
			err = static.newTransactionCalculatorError(TransactionCalculatorErrorCalcaulateTimeError)
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

func (static TransactionCalculator)ceilTimeInterval(t float64, uint float64)(v float64) {
	//起步价一个单位
	v = (math.Ceil(t/uint))
	if v == 0 {
		v = 1
	}
	return
}


type TransactionCalculatorErr struct  {
	code int
	message string
}

func (this *TransactionCalculatorErr) Error ()string{
	return this.message
}
