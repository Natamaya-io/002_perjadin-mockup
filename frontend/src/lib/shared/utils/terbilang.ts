export function terbilang(nominal: number): string {
  const bilangan = [
    '', 'Satu', 'Dua', 'Tiga', 'Empat', 'Lima', 'Enam', 'Tujuh', 'Delapan', 'Sembilan', 'Sepuluh', 'Sebelas'
  ];
  
  let temp = '';
  
  if (nominal < 12) {
    temp = ' ' + bilangan[nominal];
  } else if (nominal < 20) {
    temp = terbilang(nominal - 10) + ' Belas';
  } else if (nominal < 100) {
    temp = terbilang(Math.floor(nominal / 10)) + ' Puluh ' + terbilang(nominal % 10);
  } else if (nominal < 200) {
    temp = ' Seratus ' + terbilang(nominal - 100);
  } else if (nominal < 1000) {
    temp = terbilang(Math.floor(nominal / 100)) + ' Ratus ' + terbilang(nominal % 100);
  } else if (nominal < 2000) {
    temp = ' Seribu ' + terbilang(nominal - 1000);
  } else if (nominal < 1000000) {
    temp = terbilang(Math.floor(nominal / 1000)) + ' Ribu ' + terbilang(nominal % 1000);
  } else if (nominal < 1000000000) {
    temp = terbilang(Math.floor(nominal / 1000000)) + ' Juta ' + terbilang(nominal % 1000000);
  } else if (nominal < 1000000000000) {
    temp = terbilang(Math.floor(nominal / 1000000000)) + ' Milyar ' + terbilang(nominal % 1000000000);
  }
  
  return temp.trim();
}
