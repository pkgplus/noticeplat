package wechat

import (
	"log"
	"testing"
)

func TestSendMsg(t *testing.T) {
	openid := "odTIQ0dMoifGFIMrIoWA20G53-OA"
	tempid := "8U98v1g7PWLZ5p4jbWNSpY5dr-hhG5kVuMAUew4PHnY"
	formid := "e74d48675ff174af169df96c17289119"

	msg := NewTemplateMsg(
		openid, tempid, formid,
		[]string{"薛冰", "上班", "海尔大学", "08:43", "下班打卡"},
	)
	msg.SetEmphasis("5")
	err := SendMsg(msg)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDecrptData(t *testing.T) {
	data := "UHOh0FewgIgAFR0HkPzLmHDM6inK+GAVs0av2I3euev87/8/Kxpl7Epe3X9Cxor2K+QrU6PDha2YQyBaOwvL+6YvdBJg1HzZluTum0cGUYpk2+2IfU3VKLc6kliC8ZTvJoolKr9rG22Qclr4sTQudSLW+i9T1mxuJvDVWPk1KuieRX6EeQ0FbaqG83YWPPki8uz63IkhkyaU93sJHYGW6MQEz/Zcvc4MRW0w/G47AKKWFhML+He6emt+6iVzca01ducizemaIB/6S3my+2QM1ETR1kMkZRMf9HKlQdG6PmVmpbArSpxWOTJA51mCKruHzX0txKXJTUUn/vUnqVwHrEoGVuRRYlEMcySuxwOEVbEJAeGacX0dD167fJvMhQdqXoh47UL6Lc8MxlpW1IaI7y8bhYmw9a8rRZbzQjxw5LqwR00XP9Nmt8oWZkrwum7SXLrXKi8V+crn8Z6/wZ0C/EEdOiBxFxbFgwds0hTnkM4="
	iv := "Y6vGNrDx6JcwNo7xdJi5cg=="
	sess_key := "Ed3p6xtT7Md2vUpLQOdKZw=="

	rawdata, err := AesDecrypt(
		Base64Decode(data),
		Base64Decode(sess_key),
		Base64Decode(iv),
	)
	if err != nil {
		t.Fatal(err)
	}

	log.Printf("%s", rawdata)
}
