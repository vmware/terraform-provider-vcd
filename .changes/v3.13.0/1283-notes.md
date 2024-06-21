* Amend `TestAccVcdCatalogSharedAccess`, it failed in VCD 10.6+ as the used VDC was missing the
  `ResourceGuaranteedMemory` parameter (Flex allocation model) [GH-1283]