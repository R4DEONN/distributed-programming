import os
import sys

# Укажи здесь путь к папке с примером User-микросервиса
# Например: PROJECT_ROOT = "./user-service"
# Или запусти скрипт из той же папки, где лежит код — тогда оставь "."
PROJECT_ROOT = "."

# Расширения файлов, которые нас интересуют
INCLUDED_EXTENSIONS = {
    '.go', '.mod', '.sum',
    '.yaml', '.yml',
    '.env', '.env.example',
    'dockerfile', 'Dockerfile',
    '.json', '.toml',
    '.sql', '.sh', '.txt'
}

def should_include_file(filename: str) -> bool:
    _, ext = os.path.splitext(filename.lower())
    basename = os.path.basename(filename).lower()
    return ext in INCLUDED_EXTENSIONS or basename in {'dockerfile', 'docker-compose.yml'}

def main():
    output_lines = []
    root_abs = os.path.abspath(PROJECT_ROOT)

    for dirpath, dirnames, filenames in os.walk(root_abs):
        # Пропускаем служебные папки
        dirnames[:] = [d for d in dirnames if d not in ('.git', '__pycache__', 'node_modules', 'dist', 'build')]
        for filename in filenames:
            filepath = os.path.join(dirpath, filename)
            relpath = os.path.relpath(filepath, root_abs)

            if should_include_file(filename):
                try:
                    with open(filepath, 'r', encoding='utf-8') as f:
                        content = f.read()
                except Exception as e:
                    content = f"[ERROR: Could not read file: {e}]"

                output_lines.append(f"=== FILE: {relpath} ===")
                output_lines.append(content)
                output_lines.append("")  # пустая строка между файлами

    with open("project_dump.txt", "w", encoding="utf-8") as out:
        out.write("\n".join(output_lines))

    print("✅ Сборка завершена! Результат сохранён в 'project_dump.txt'")
    print("📁 Файл находится в той же папке, где запущён скрипт.")

if __name__ == "__main__":
    main()