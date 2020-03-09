//
// Copyright (c) 2018
// Circutor
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"strconv"

	"github.com/edgexfoundry/edgex-go/pkg/models"
)

var DexmaParameterTable = []struct {
	name    string
	code    string
	codenum int
}{
	{"POWER", "401", 401},       //	401	Active	Power	W
	{"EACTIVE", "402", 402},     //	402	Active Energy	kWh
	{"IRPOWER", "403", 403},     //	403	Inductive	Reactive	Power	VAr
	{"IRENERGY", "404", 404},    //	404	Reactive	Energy	kVArh
	{"VOLTAGE", "405", 405},     //	405	Voltage	V
	{"CURRENT", "406", 406},     //	406	Current	A
	{"CRPOWER", "407", 407},     //	407	Capacitive	Reactive	Power	VAr
	{"CRENERGY", "408", 408},    //	408	Capacitive	Reactive	Energy	kVArh
	{"APPOWER", "409", 409},     //	409	Apparent	Power	VA
	{"APENERGY", "410", 410},    //	410	Apparent	Energy	kVAh
	{"COSPHY", "411", 411},      //	411	Cos	Phi	-
	{"PF", "412", 412},          //	412	Power	Factor	-
	{"NCURRENT", "413", 413},    //	413	Neutral	Current	A
	{"FREQ", "414", 414},        //	414	Frequency	Hz
	{"IAENERGY", "415", 415},    //	415	Intv.	Active	Energy	kWh
	{"IIRENERGY", "416", 416},   //	416	Intv.	Inductive	R.	Energy	kVArh
	{"IAPENERGY", "417", 417},   //	417	Intv.	Apparent	Energy	kVAh
	{"MAXDEMAND", "418", 418},   //	418	Intv.	Maximum	demand	W
	{"GASVOLUME", "419", 419},   //	419	Gas	volume	m3
	{"GASENERGY", "420", 420},   //	420	Gas	Energy	kWh
	{"GASVOLN", "421", 421},     //	421	Gas	volume	normalized	Nm3
	{"THDV", "422", 422},        //	422	Voltage	distortion	(THD-total	harmonic	distorsion)	%
	{"THDC", "423", 423},        //	423	Current	distortion	(THD-total	harmonic	distorsion)	%
	{"IMAXVOLTAGE", "424", 424}, //	424	Intv.	Maximum	voltage	V
	{"IMAXCURRENT", "425", 425}, //	425	Intv.	Maximum	current	A
	{"AVGCURRENT", "426", 426},  //	426	Average	Current	A
	{"PASEENERGY", "427", 427},  //	427	Night	Passive	Consumption	kWh
	{"BASEENERGY", "428", 428},  //	428	Permanent	Passive	Consumption	kWh
	{"EACTIVEABS", "429", 429},  //	429	Absolute	reading	electricity	meter	kWh

}

// newHTTPDexmaSender - create http dexma sender
func newHTTPDexmaSender(addr models.Addressable) sender {
	sender := httpSender{
		url:    addr.Protocol + "://" + addr.Address + ":" + strconv.Itoa(addr.Port) + "/" + addr.Topic + "?source_key=" + addr.User + "&dexcell_source_token=" + addr.Password,
		method: addr.HTTPMethod,
	}
	return sender
}

func transformDexmaParam(name string) int {

	for i := 0; i < len(DexmaParameterTable); i++ {
		if name == DexmaParameterTable[i].name || name == DexmaParameterTable[i].code {
			return DexmaParameterTable[i].codenum
		}
	}
	return 0
}
