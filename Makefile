.PHONY: dev down logs build stop watch

dev:
	@echo "Starting development environment (SOTA: Bind Mounts & Native HMR)..."
	doppler run -- podman compose up -d

build:
	podman compose build backend
	podman compose build frontend

stop:
	podman compose stop

down:
	doppler run -- podman compose down -v

logs:
	podman compose logs -f

watch:
	@echo "⚠️ TARGET 'watch' KINI USANG (DEPRECATED) ⚠️"
	@echo "Sistem menggunakan arsitektur SOTA (Vite HMR & Go Air)."
	@echo "Anda tidak perlu melakukan restart manual. Kode akan ter-update otomatis dalam hitungan milidetik saat Anda melakukan Save."
	@echo "Menjalankan target 'dev'..."
	$(MAKE) dev
