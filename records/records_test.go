package records

import (
	"fmt"
	"testing"
)

func TestCombineIPv6PrefixAndInterfaceID(t *testing.T) {
	tests := []struct {
		prefix   string
		ipv6     string
		expected string
	}{
		{"2001:db8::", "::1234:5678:90ab:cdef:0123", "2001:db8::5678:90ab:cdef:123"},
		{"2001:db8::", "fe80:efef:db8:1234:5678:90ab:cdef:0123", "2001:db8::5678:90ab:cdef:123"},
		{"2001:efef:db8:1234::101", "fe80:efef:db8:1234:5678:90ab:cdef:0123", "2001:efef:db8:1234:5678:90ab:cdef:123"},
	}

	for index, testcase := range tests {
		t.Run(fmt.Sprintf("TestCase%d", index+1), func(t *testing.T) {
			result := combineIPv6PrefixAndInterfaceID(testcase.prefix, testcase.ipv6)
			if result != testcase.expected {
				t.Errorf("expected: %s, got: %s", testcase.expected, result)
			}
			// assert.Assert(testcase.expected == result, t)
		})
	}
}

func TestGetSubAndRootDomain(t *testing.T) {
	tests := []struct {
		fqdn         string
		expectedSub  string
		expectedRoot string
	}{
		{"sub.example.com", "sub", "example.com"},
		{"example.com", "", "example.com"},
		{"sub.sub.example.com", "sub.sub", "example.com"},
		{"*.example.com", "*", "example.com"},
	}

	for _, testcase := range tests {
		t.Run(testcase.fqdn, func(t *testing.T) {
			sub, root := getSubAndRootDomain(testcase.fqdn)
			if sub != testcase.expectedSub || root != testcase.expectedRoot {
				t.Errorf("sub: %s, root: %s", sub, root)
			}
			// assert.Assert(testcase.expectedSub == sub, t)
			// assert.Assert(testcase.expectedRoot == root, t)
		})
	}
}

func TestIsFQDNValid(t *testing.T) {
	tests := []struct {
		fqdn     string
		expected bool
	}{
		{"example.de", true},
		{"sub.example.com", true},
		{"*.example.com", true},
		{"example", false},
		{"example.c", false},
		{"example..com", false},
	}

	for _, testcase := range tests {
		t.Run(testcase.fqdn, func(t *testing.T) {
			result := isFQDNValid(testcase.fqdn)
			if result != testcase.expected {
				t.Errorf("expected: %t, got: %t", testcase.expected, result)
			}
			// assert.Assert(testcase.expected == result, t)
		})
	}
}
