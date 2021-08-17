package types

import (
	"fmt"
	"strings"
)

type Permission string

const ReadSecret Permission = "read:secrets"
const SetSecret Permission = "set:secrets"
const DeleteSecret Permission = "delete:secrets"
const DestroySecret Permission = "destroy:secrets"

const ReadKey Permission = "read:keys"
const SetKey Permission = "set:keys"
const DeleteKey Permission = "delete:keys"
const DestroyKey Permission = "destroy:keys"
const SignKey Permission = "sign:keys"
const EncryptKey Permission = "encrypt:keys"

const ReadEth1 Permission = "read:eth1accounts"
const SetEth1 Permission = "set:eth1accounts"
const DeleteEth1 Permission = "delete:eth1accounts"
const DestroyEth1 Permission = "destroy:eth1accounts"
const SignEth1 Permission = "sign:eth1accounts"
const EncryptEth1 Permission = "encrypt:eth1accounts"

func ListPermissions() []Permission {
	return []Permission{
		ReadSecret,
		SetSecret,
		DeleteSecret,
		DestroySecret,
		ReadKey,
		SetKey,
		DeleteKey,
		DestroyKey,
		SignKey,
		EncryptKey,
		ReadEth1,
		SetEth1,
		DeleteEth1,
		DestroyEth1,
		SignEth1,
		EncryptEth1,
	}
}

func ListWildcardPermission(p string) []Permission {
	all := ListPermissions()
	parts := strings.Split(p, ":")
	action, resource := parts[0], parts[1]
	if action == "*" && resource == "*" {
		return all
	}

	included := []Permission{}
	for _, ip := range all {
		if action == "*" && strings.Contains(string(ip), fmt.Sprintf(":%s", resource)) {
			included = append(included, ip)
		}
		if resource == "*" && strings.Contains(string(ip), fmt.Sprintf("%s:", action)) {
			included = append(included, ip)
		}
	}

	return included
}