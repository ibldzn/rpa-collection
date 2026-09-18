# Inquiry

Endpoint: GET /pinjaman/updateManualKolek/pembuatan/cari

Query:
- norekening=3000010000000010

Response:
```json
{
  "data": {
    "result": {
      "nocif": "00100000432",
      "namanasabah": "Ratna Juwita",
      "jenisnasabah": "Perorangan",
      "tglbukacif": {
        "date": "2015-04-18 00:00:00.000000",
        "timezone_type": 3,
        "timezone": "UTC"
      },
      "status_dokumen": "Aktif",
      "dataalamat_ktp_alamat1": "Kp Cibitung Sebrang RT 02/09",
      "dataalamat_ktp_alamat2": null,
      "dataalamat_ktp_rt": "000",
      "dataalamat_ktp_rw": "000",
      "dataalamat_ktp_kelurahan": "Cimuning",
      "dataalamat_ktp_kecamatan": "Mustika Jaya",
      "dataalamat_ktp_kota": "0102",
      "dataalamat_ktp_propinsi": "0",
      "dataalamat_ktp_kodepos": "17155",
      "dataktp_nik": null,
      "dataktp_nama": "Ratna Juwita",
      "dataktp_tempatlahir": null,
      "dataktp_jeniskelamin": null,
      "dataktp_golongandarah": null,
      "dataktp_alamat": null,
      "dataktp_agama": null,
      "dataktp_statusperkawinan": null,
      "dataktp_pekerjaan": null,
      "dataktp_kewarganegaraan": null,
      "dataktp_tglberlaku": null,
      "dataktp_tempatterbit": null,
      "dataktp_tglterbit": null,
      "dataktp_tgllahir": {
        "date": "1969-09-19 00:00:00.000000",
        "timezone_type": 3,
        "timezone": "UTC"
      },
      "dataktp_berlakuseumurhidup": null,
      "locationname": "Kantor Pusat Operasional",
      "rec_dibuat_oleh": "IT",
      "norekening": "3000010000000010",
      "noalt": "0130101394",
      "statusrekening": "Aktif",
      "nopk": "PL001000073837",
      "appdate": {
        "date": "2026-08-27 00:00:00.000000",
        "timezone_type": 3,
        "timezone": "UTC"
      },
      "tglpk": {
        "date": "2013-03-18 00:00:00.000000",
        "timezone_type": 3,
        "timezone": "UTC"
      },
      "currency": "IDR",
      "plafondlimit": 30000000,
      "longgartarik": 21800000,
      "terpakai": 30000000,
      "periode": "2013-03-18 - 2025-10-12",
      "revolving": null,
      "tglakhirpencairan_pk": {
        "date": "2025-10-12 00:00:00.000000",
        "timezone_type": 3,
        "timezone": "UTC"
      },
      "keterangan": "-",
      "referensi": "01.301.01394",
      "officerpk": "IT",
      "datarekening": {
        "id": "3000010000000010",
        "nopk": "PL001000073837",
        "namanasabah": "Ratna Juwita",
        "nama": "301 - Kredit Pegawai Aktif",
        "status_dokumen": "Aktif",
        "jenispinjaman": "Kredit Angsuran",
        "sukubunga": 24,
        "perubahansukubunga": "Bunga Fix",
        "periode": "2013-03-18 - 2018-03-18",
        "jangkawaktu": "60 Month",
        "currency": "IDR",
        "kolekbimanual": 0,
        "kolekbprmanual": 2,
        "kolekbiauto": 5,
        "kolekbprauto": 2,
        "saldopinjaman": 0,
        "tunggakanbunga": 16650000,
        "tunggakandenda": 0,
        "dpdview": "4423 Hari",
        "dpd": 4423,
        "kolekbi": 5,
        "kolekbpr": 2,
        "updatekolekbi": "Automatic",
        "updatekolekbpr": "Manual",
        "totalnilaijaminan": 0,
        "jmlpokok_pinjaman": 30000000,
        "totalassetvalue": 0,
        "totalcollateralvalue": 0
      }
    }
  },
  "status": "ok"
}
```
```
```

# Update Kolek
Endpoint: POST /pinjaman/updateManualKolek/pembuatan/pinjaman
Header:
- Accept: application/json

