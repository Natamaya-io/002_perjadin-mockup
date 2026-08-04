-- Sengaja tidak menghapus data master GUP saat rollback.
-- Penghapusan otomatis dapat merusak transaksi GUP yang sudah memakai data ini.
SELECT 1;