* Internal functions `lockParentEdgeGtw`, `unLockParentEdgeGtw`, `lockEdgeGateway`,
  `unlockEdgeGateway` were converted to use just their ID for lock key instead of full path
  `org:vdc:edge_id`. This is done because paths for VDC and VDC Groups can differ, but UUID is
  unique so it makes it simpler to manage [GH-YYY]