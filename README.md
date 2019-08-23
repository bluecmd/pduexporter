# pduexporter

Usage: `./pduexporter --host 192.168.0.216`

Example output:

```
# HELP pdu_current PDU Total Current in Amperes
# TYPE pdu_current gauge
pdu_current 1.7
# HELP pdu_outlet_status PDU Outlet boolean status
# TYPE pdu_outlet_status gauge
pdu_outlet_status{index="0",name="Ascreen"} 1
pdu_outlet_status{index="1",name="Bzotac"} 1
pdu_outlet_status{index="2",name="CsanB1"} 0
pdu_outlet_status{index="3",name="DsanA1"} 1
pdu_outlet_status{index="4",name="Eempty"} 1
pdu_outlet_status{index="5",name="Fempty"} 1
pdu_outlet_status{index="6",name="Gocp"} 1
pdu_outlet_status{index="7",name="Harista1"} 1
# HELP pdu_system PDU System tracker
# TYPE pdu_system counter
pdu_system{firmware="s4.82-091012-1cb08s",location="KekColo",model="SWH-1023J-08N1",name="PDU A power"} 1
# HELP pdu_voltage PDU Voltage in Volts
# TYPE pdu_voltage gauge
pdu_voltage 230
``
