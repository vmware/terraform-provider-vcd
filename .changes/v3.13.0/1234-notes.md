* Improve log traceability by sending `X-VMWARE-VCLOUD-CLIENT-REQUEST-ID` header in requests. The
  header will be formatted in such format `162-2024-04-11-08-41-34-171-` where the first number
  (`162`) is the API call sequence number in the life of that particular process followed by a
  hyphen separated date time with millisecond precision (`2024-04-11-08-41-34-171` for April 11th of
  year 2024 at time 08:41:34.171). The trailing hyphen `-` is here to separate response header
  `X-Vmware-Vcloud-Request-Id` suffix with double hyphen
  `162-2024-04-11-08-41-34-171--40d78874-27a3-4cad-bd43-2764f557226b` [GH-1234]
