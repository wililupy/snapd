// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2018 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package builtin_test

import (
	. "gopkg.in/check.v1"

	"github.com/snapcore/snapd/interfaces"
	"github.com/snapcore/snapd/interfaces/apparmor"
	"github.com/snapcore/snapd/interfaces/builtin"
	"github.com/snapcore/snapd/snap"
	"github.com/snapcore/snapd/snap/snaptest"
	"github.com/snapcore/snapd/testutil"
)

type HugepagesControlSuite struct {
	iface    interfaces.Interface
	slotInfo *snap.SlotInfo
	slot     *interfaces.ConnectedSlot
	plugInfo *snap.PlugInfo
	plug     *interfaces.ConnectedPlug
}

var _ = Suite(&HugepagesControlSuite{
	iface: builtin.MustInterface("hugepages-control"),
})

func (s *HugepagesControlSuite) SetUpTest(c *C) {
	const producerYaml = `name: core
version: 0
type: os
slots:
  hugepages-control:
`
	info := snaptest.MockInfo(c, producerYaml, nil)
	s.slotInfo = info.Slots["hugepages-control"]
	s.slot = interfaces.NewConnectedSlot(s.slotInfo, nil)

	const consumerYaml = `name: consumer
version: 0
apps:
 app:
  plugs: [hugepages-control]
`
	info = snaptest.MockInfo(c, consumerYaml, nil)
	s.plugInfo = info.Plugs["hugepages-control"]
	s.plug = interfaces.NewConnectedPlug(s.plugInfo, nil)
}

func (s *HugepagesControlSuite) TestName(c *C) {
	c.Assert(s.iface.Name(), Equals, "hugepages-control")
}

func (s *HugepagesControlSuite) TestSanitizeSlot(c *C) {
	c.Assert(interfaces.BeforePrepareSlot(s.iface, s.slotInfo), IsNil)
	slot := &snap.SlotInfo{
		Snap:      &snap.Info{SuggestedName: "some-snap"},
		Name:      "hugepages-control",
		Interface: "hugepages-control",
	}
	c.Assert(interfaces.BeforePrepareSlot(s.iface, slot), ErrorMatches,
		"hugepages-control slots are reserved for the core snap")
}

func (s *HugepagesControlSuite) TestSanitizePlug(c *C) {
	c.Assert(interfaces.BeforePreparePlug(s.iface, s.plugInfo), IsNil)
}

func (s *HugepagesControlSuite) TestAppArmorSpec(c *C) {
	spec := &apparmor.Specification{}
	c.Assert(spec.AddConnectedPlug(s.iface, s.plug, s.slot), IsNil)
	c.Assert(spec.SecurityTags(), DeepEquals, []string{"snap.consumer.app"})
	c.Assert(spec.SnippetForTag("snap.consumer.app"), testutil.Contains, "/sys/kernel/mm/hugepages/* rw,")
}

func (s *HugepagesControlSuite) TestStaticInfo(c *C) {
	si := interfaces.StaticInfoOf(s.iface)
	c.Assert(si.ImplicitOnCore, Equals, true)
	c.Assert(si.ImplicitOnClassic, Equals, true)
	c.Assert(si.Summary, Equals, "allows using hugepages")
	c.Assert(si.BaseDeclarationSlots, testutil.Contains, "hugepages-control")
}

func (s *HugepagesControlSuite) TestAutoConnect(c *C) {
	// FIXME: fix AutoConnect methods to use ConnectedPlug/Slot
	c.Assert(s.iface.AutoConnect(&interfaces.Plug{PlugInfo: s.plugInfo}, &interfaces.Slot{SlotInfo: s.slotInfo}), Equals, true)
}

func (s *HugepagesControlSuite) TestInterfaces(c *C) {
	c.Check(builtin.Interfaces(), testutil.DeepContains, s.iface)
}
