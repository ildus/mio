package main

import (
	"github.com/paypal/gatt"
	validator "gopkg.in/go-playground/validator.v8"
	"log"
	"time"
)

type Gid uint16
type LGid string

// Standart
const (
	SERVICE_BATTERY Gid = 0x180F
	CHAR_BATTERY    Gid = 0x2A19 //read notify, one byte

	SERVICE_HEART_RATE Gid = 0x180D
	CHAR_HEART_RATE    Gid = 0x2A37 //notify
	CHAR_BODY_SENSOR   Gid = 0x2A38 //read
)

// Not standart UUIDs
const (
	SERVICE_MIO_SPORT                 = "6C721838-5BF1-4F64-9170-381C08EC57EE"
	CHARACTERISTIC_MIO_SPORT_MSG      = "6C722A80-5BF1-4F64-9170-381C08EC57EE" //read write 0001
	CHARACTERISTIC_MIO_SPORT_UNKNOWN  = "6C722A81-5BF1-4F64-9170-381C08EC57EE" //read write 00
	CHARACTERISTIC_MIO_SPORT_MSG_RESP = "6C722A82-5BF1-4F64-9170-381C08EC57EE" //read notify
	CHARACTERISTIC_MIO_SENSOR         = "6C722A83-5BF1-4F64-9170-381C08EC57EE" //read notify
	CHARACTERISTIC_MIO_RECORD         = "6C722A84-5BF1-4F64-9170-381C08EC57EE" //read notify
)

// commands
const (
	CMD_TYPE_NONE = iota
	CMD_TYPE_HR_SET
	CMD_TYPE_HR_GET
	CMD_TYPE_BIKE_SET
	CMD_TYPE_BIKE_GET
	CMD_TYPE_APPTYPE_GET
	CMD_TYPE_APPTYPE_SET
	CMD_TYPE_USERINFO_GET
	CMD_TYPE_USERINFO_SET
	CMD_TYPE_NAME_GET
	CMD_TYPE_NAME_SET
	CMD_TYPE_RTC_GET
	CMD_TYPE_RTC_SET
	CMD_TYPE_RUN_CMD
	CMD_TYPE_SEND_GPS_DATA
	CMD_TYPE_DISPLAY_GET
	CMD_TYPE_DISPLAY_SET
	CMD_TYPE_DAILYGOAL_GET
	CMD_TYPE_DAILYGOAL_SET
	CMD_TYPE_DEVICE_STATUS_GET
	CMD_TYPE_TODAY_ADL_RECORD_GET
	CMD_TYPE_RECORD_GET
	CMD_TYPE_RECORD_DELETE
	CMD_TYPE_SESSION_GET
	CMD_TYPE_LINK_CUST_CMD
	CMD_TYPE_LINK_ENTER_DFUMODE
	CMD_TYPE_ALPHA2_ENTER_DFUMODE
	CMD_TYPE_LINK_UPDATE
	CMD_TYPE_ALPHA2_UPDATE
	CMD_TYPE_STRIDE_CALI_GET
	CMD_TYPE_STRIDE_CALI_SET
	CMD_TYPE_FACTORY_DEFAULT
	CMD_TYPE_SWING_ARM_GET
	CMD_TYPE_SWING_ARM_SET
	CMD_TYPE_VELO_DEVICE_STATUS_GET
	CMD_TYPE_VELO_MEM_RECORD_GET
	CMD_TYPE_VELO_MEM_SESSION_GET
	CMD_TYPE_VELO_MEM_RECORD_DEL
	CMD_TYPE_LINK_MOBILE_NOTIFICATION
	CMD_TYPE_LINK_MOBILE_MSG_ALERT
	CMD_TYPE_LINK_MOBILE_EMAIL_ALERT
	CMD_TYPE_LINK_MOBILE_PHONE_ALERT
	CMD_TYPE_SLEEP_RECORD_GET
	CMD_TYPE_SLEEP_RECORD_DELETE
	CMD_TYPE_SLEEP_RECORD_CURHOUR
	CMD_TYPE_DEVICE_OPTION_GET
	CMD_TYPE_DEVICE_OPTION_SET
)

const (
	CMD_RUN_STREAM_MODE_DISABLE = iota
	CMD_RUN_STREAM_MODE_ENABLE
	CMD_RUN_GPS_MODE_DISABLE
	CMD_RUN_GPS_MODE_ENABLE
	CMD_RUN_RESET_TODAY_ADL_DATA
	CMD_RUN_STEP_DATA_NOTIFY_DISABLE
	CMD_RUN_STEP_DATA_NOTIFY_ENABLE
	CMD_RUN_AIRPLANE_MODE_ENABLE
	CMD_RUN_MEM_CLEAR
	CMD_RUN_USERDATA_BACKUP
	CMD_RUN_ETS_NOTIFICATION_DISABLE //ExeTimeSyncData
	CMD_RUN_ETS_NOTIFICATION_ENABLE
	CMD_RUN_ETS_STARTTIMER //ExeTimerSyncCmd
	CMD_RUN_ETS_STOPTIMER
	CMD_RUN_ETS_TAKELAP
	CMD_RUN_ETS_RESEND_LAP
	CMD_RUN_ETS_FINISH
	CMD_RUN_SLEEP_MODE_DEACTIVATE
	CMD_RUN_SLEEP_MODE_ACTIVATE
	CMD_RUN_REST_HR_TAKE_MEASUREMENT
	CMD_RUN_REST_HR_STOP_MEASUREMENT
	CMD_RUN_REST_HR_SEND_MEASUREMENT
	CMD_RUN_ACT_MEM_CLEAR
	CMD_RUN_ADL_MEM_CLEAR
)

type UserInfo struct {
	Gender             byte `validate:"min=0,max=1"`
	UnitType           byte `validate:"min=0,max=1"`
	HRDisplayType      byte `validate:"min=0,max=1"`
	DisplayOrientation byte `validate:"min=0,max=1"`
	WODisplayMode      byte `validate:"min=0,max=1"`
	ADLGoalCal         byte `validate:"min=0,max=1"`
	WORecording        byte `validate:"min=0,max=1"`
	HRAutoAdj          byte `validate:"min=0,max=1"`
	Birthday           time.Time
	BodyWeight         byte `validate:"min=20,max=200"`
	BodyHeight         byte `validate:"min=69,max=231"`
	ResetHR            byte `validate:"min=30,max=140,ltfield=MaxHR"`
	MaxHR              byte `validate:"min=80,max=220"`
}

// userinfo bitmask values
const (
	FLAG_GENDER = 1 << iota
	FLAG_UNIT_TYPE
	FLAG_DISPLAY_TYPE_HR
	FLAG_DISPLAY_ORIENTATION
	FLAG_DISPLAY_MODE_WO
	FLAG_GOAL_ADL
	FLAG_RECORDING_WO
	FLAG_ADJ_HR
)

var (
	validate *validator.Validate
)

func (i Gid) asUUID() []gatt.UUID {
	return []gatt.UUID{gatt.UUID16(uint16(i))}
}

func (i LGid) asUUID() []gatt.UUID {
	return []gatt.UUID{gatt.MustParseUUID(string(i))}
}

func initValidator() {
	config := &validator.Config{TagName: "validate"}
	validate = validator.New(config)
}

func (u *UserInfo) Pack() *[]byte {
	var flags byte = 0

	errs := validate.Struct(u)
	if errs != nil {
		log.Println("Validation error", errs)
		return nil
	}
	data := make([]byte, 10)

	if u.Gender == 1 {
		flags |= FLAG_GENDER
	}

	if u.UnitType == 1 {
		flags |= FLAG_UNIT_TYPE
	}

	if u.HRDisplayType == 1 {
		flags |= FLAG_DISPLAY_TYPE_HR
	}

	if u.DisplayOrientation == 1 {
		flags |= FLAG_DISPLAY_ORIENTATION
	}

	if u.WODisplayMode == 1 {
		flags |= FLAG_DISPLAY_MODE_WO
	}

	if u.ADLGoalCal == 1 {
		flags |= FLAG_GOAL_ADL
	}

	if u.WORecording == 1 {
		flags |= FLAG_RECORDING_WO
	}

	if u.HRAutoAdj == 1 {
		flags |= FLAG_ADJ_HR
	}

	data[0] = 8 //i don't know what it's mean
	data[1] = 0 //msg_user_settings_set
	data[2] = flags
	data[3] = byte(u.Birthday.Day())
	data[4] = byte(u.Birthday.Month())
	data[5] = byte(u.Birthday.Year())
	data[6] = u.BodyWeight
	data[7] = u.BodyHeight
	data[8] = u.ResetHR
	data[9] = u.MaxHR
	return &data
}

func (u *UserInfo) Unpack(data *[]byte) {

}

/*
Service: 180d (Heart Rate)
  Characteristic  2a37 (Heart Rate Measurement)
    properties    notify
  Descriptor      2902 (Client Characteristic Configuration)
    value         0000 | "\x00\x00"
  Characteristic  2a38 (Body Sensor Location)
    properties    read
    value         02 | "\x02"
2015/12/28 02:00:31 Unhandled event: xpc.Dict{"kCBMsgId":81, "kCBMsgArgs":xpc.Dict{"kCBMsgArgDeviceUUID":xpc.UUID{0xcf, 0xd4, 0x72, 0xc6, 0x31, 0xbc, 0x4d, 0x8a, 0x8c, 0xe8, 0x3c, 0x16, 0x6c, 0xf1, 0x21, 0xea}, "kCBMsgArgConnectionInterval":498, "kCBMsgArgConnectionLatency":0, "kCBMsgArgSupervisionTimeout":500}}

Service: 180f (Battery Service)
  Characteristic  2a19 (Battery Level)
    properties    read notify
    value         3a | ":"
  Descriptor      2902 (Client Characteristic Configuration)
    value         0000 | "\x00\x00"

Service: 6c7218265bf14f649170381c08ec57ee
  Characteristic  6c722a0a5bf14f649170381c08ec57ee
    properties    read write
    value         1100000000000000000000000000000000000011 | "\x11\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x11"
  Characteristic  6c722a0b5bf14f649170381c08ec57ee
    properties    read write
    value         2200000000000000000000000000000000000022 | "\"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\""
  Characteristic  6c722a0c5bf14f649170381c08ec57ee
    properties    read notify
    value         3300000000000000000000000000000000000033 | "3\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x003"
  Descriptor      2902 (Client Characteristic Configuration)
    value         0000 | "\x00\x00"
2015/12/28 02:00:48 Unhandled event: xpc.Dict{"kCBMsgId":81, "kCBMsgArgs":xpc.Dict{"kCBMsgArgDeviceUUID":xpc.UUID{0xcf, 0xd4, 0x72, 0xc6, 0x31, 0xbc, 0x4d, 0x8a, 0x8c, 0xe8, 0x3c, 0x16, 0x6c, 0xf1, 0x21, 0xea}, "kCBMsgArgConnectionInterval":498, "kCBMsgArgConnectionLatency":0, "kCBMsgArgSupervisionTimeout":500}}

Service: 6c7218385bf14f649170381c08ec57ee
  Characteristic  6c722a805bf14f649170381c08ec57ee
    properties    read write
    value         0001 | "\x00\x01"
  Characteristic  6c722a815bf14f649170381c08ec57ee
    properties    read write
    value         00 | "\x00"
  Characteristic  6c722a825bf14f649170381c08ec57ee
    properties    read notify
    value         0b8001008b0060738699ac7194 | "\v\x80\x01\x00\x8b\x00`s\x86\x99\xacq\x94"
  Descriptor      2902 (Client Characteristic Configuration)
    value         0000 | "\x00\x00"
  Characteristic  6c722a835bf14f649170381c08ec57ee
    properties    read notify
    value         00 | "\x00"
  Descriptor      2902 (Client Characteristic Configuration)
    value         0000 | "\x00\x00"
  Characteristic  6c722a845bf14f649170381c08ec57ee
    properties    read notify
    value         04a5a4a4a4a4bfb5 | "\x04\xa5\xa4\xa4\xa4\xa4\xbf\xb5"
  Descriptor      2902 (Client Characteristic Configuration)
    value         0000 | "\x00\x00"

Service: 180a (Device Information)
  Characteristic  2a29 (Manufacturer Name String)
    properties    read
    value         4d494f20474c4f42414c | "MIO GLOBAL"
  Characteristic  2a24 (Model Number String)
    properties    read
    value         4655534520353950 | "FUSE 59P"
  Characteristic  2a25 (Serial Number String)
    properties    read
    value         3830303042343937 | "8000B497"
  Characteristic  2a27 (Hardware Revision String)
    properties    read
    value         3030313231 | "00121"
  Characteristic  2a26 (Firmware Revision String)
    properties    read
    value         30312e3138 | "01.18"
  Characteristic  2a28 (Software Revision String)
    properties    read
    value         30312e30392e3035 | "01.09.05"
  Characteristic  2a23 (System ID)
    properties    read
    value         97b40080f01acc1f3dc8 | "\x97\xb4\x00\x80\xf0\x1a\xcc\x1f=\xc8"
*/
