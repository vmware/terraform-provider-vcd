* Amend the test `TestAccVcdRdeDuplicate` so it doesn't fail on VCD 10.6+. Since this version, whenever a RDE is created
  in a tenant by the System Administrator, the owner is not `"administrator"` anymore, but `"system"` [GH-1278]
