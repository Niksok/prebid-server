package privacy

import (
	"testing"

	"github.com/mxmCherry/openrtb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAny(t *testing.T) {
	testCases := []struct {
		enforcement Enforcement
		expected    bool
		description string
	}{
		{
			description: "All False",
			enforcement: Enforcement{
				CCPA:  false,
				COPPA: false,
				GDPR:  false,
			},
			expected: false,
		},
		{
			description: "All True",
			enforcement: Enforcement{
				CCPA:  true,
				COPPA: true,
				GDPR:  true,
			},
			expected: true,
		},
		{
			description: "Mixed",
			enforcement: Enforcement{
				CCPA:  false,
				COPPA: true,
				GDPR:  false,
			},
			expected: true,
		},
	}

	for _, test := range testCases {
		result := test.enforcement.Any()
		assert.Equal(t, test.expected, result, test.description)
	}
}

func TestApply(t *testing.T) {
	testCases := []struct {
		description             string
		enforcement             Enforcement
		expectedDeviceIPv6      ScrubStrategyIPV6
		expectedDeviceGeo       ScrubStrategyGeo
		expectedUserDemographic ScrubStrategyDemographic
		expectedUserGeo         ScrubStrategyGeo
	}{
		{
			description: "All Enforced",
			enforcement: Enforcement{
				CCPA:  true,
				COPPA: true,
				GDPR:  true,
			},
			expectedDeviceIPv6:      ScrubStrategyIPV6Lowest32,
			expectedDeviceGeo:       ScrubStrategyGeoFull,
			expectedUserDemographic: ScrubStrategyDemographicAgeAndGender,
			expectedUserGeo:         ScrubStrategyGeoFull,
		},
		{
			description: "CCPA Only",
			enforcement: Enforcement{
				CCPA:  true,
				COPPA: false,
				GDPR:  false,
			},
			expectedDeviceIPv6:      ScrubStrategyIPV6Lowest16,
			expectedDeviceGeo:       ScrubStrategyGeoReducedPrecision,
			expectedUserDemographic: ScrubStrategyDemographicNone,
			expectedUserGeo:         ScrubStrategyGeoReducedPrecision,
		},
		{
			description: "COPPA Only",
			enforcement: Enforcement{
				CCPA:  false,
				COPPA: true,
				GDPR:  false,
			},
			expectedDeviceIPv6:      ScrubStrategyIPV6Lowest32,
			expectedDeviceGeo:       ScrubStrategyGeoFull,
			expectedUserDemographic: ScrubStrategyDemographicAgeAndGender,
			expectedUserGeo:         ScrubStrategyGeoFull,
		},
		{
			description: "GDPR Only",
			enforcement: Enforcement{
				CCPA:  false,
				COPPA: false,
				GDPR:  true,
			},
			expectedDeviceIPv6:      ScrubStrategyIPV6Lowest16,
			expectedDeviceGeo:       ScrubStrategyGeoReducedPrecision,
			expectedUserDemographic: ScrubStrategyDemographicNone,
			expectedUserGeo:         ScrubStrategyGeoReducedPrecision,
		},
	}

	for _, test := range testCases {
		req := &openrtb.BidRequest{
			Device: &openrtb.Device{},
			User:   &openrtb.User{},
		}
		replacedDevice := &openrtb.Device{}
		replacedUser := &openrtb.User{}

		m := &mockScrubber{}
		m.On("ScrubDevice", req.Device, test.expectedDeviceIPv6, test.expectedDeviceGeo).Return(replacedDevice).Once()
		m.On("ScrubUser", req.User, test.expectedUserDemographic, test.expectedUserGeo).Return(replacedUser).Once()

		test.enforcement.apply(req, m)

		m.AssertExpectations(t)
		assert.Same(t, replacedDevice, req.Device, "Device")
		assert.Same(t, replacedUser, req.User, "User")
	}
}

func TestApplyNoneApplicable(t *testing.T) {
	req := &openrtb.BidRequest{}

	m := &mockScrubber{}

	enforcement := Enforcement{
		CCPA:  false,
		COPPA: false,
		GDPR:  false,
	}
	enforcement.apply(req, m)

	m.AssertNotCalled(t, "ScrubDevice")
	m.AssertNotCalled(t, "ScrubUser")
}

func TestApplyNil(t *testing.T) {
	m := &mockScrubber{}

	enforcement := Enforcement{}
	enforcement.apply(nil, m)

	m.AssertNotCalled(t, "ScrubDevice")
	m.AssertNotCalled(t, "ScrubUser")
}

type mockScrubber struct {
	mock.Mock
}

func (m *mockScrubber) ScrubDevice(device *openrtb.Device, ipv6 ScrubStrategyIPV6, geo ScrubStrategyGeo) *openrtb.Device {
	args := m.Called(device, ipv6, geo)
	return args.Get(0).(*openrtb.Device)
}

func (m *mockScrubber) ScrubUser(user *openrtb.User, demographic ScrubStrategyDemographic, geo ScrubStrategyGeo) *openrtb.User {
	args := m.Called(user, demographic, geo)
	return args.Get(0).(*openrtb.User)
}
