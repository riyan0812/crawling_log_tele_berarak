from telethon.sync import TelegramClient
from telethon import events
import os
import requests
import asyncio
import zipfile
import rarfile
import py7zr

# API dan nomor telepon
api_id = '21512835'
api_hash = 'e9673810b4af69b69db23810f7034af7'
phone_number = '6281996809571'

# Folder tempat file akan disimpan
download_folder = '/harvest2'

# Pastikan folder tersedia
if not os.path.exists(download_folder):
    os.makedirs(download_folder)

client = TelegramClient('session_name', api_id, api_hash)

# URL webhook Discord
discord_webhook_url = 'https://discord.com/api/webhooks/1329896555200647208/DHux2YZ0C-hkYR8sw2toLvoEpaT_Wej4YcqCp7lCGyEPXUraPdE0k-Mv0d3dtGqESHrg'

# Fungsi kirim notifikasi ke Discord
def send_discord_notification(message):
    data = {"content": message}
    try:
        response = requests.post(discord_webhook_url, json=data)
        if response.status_code in [200, 204]:
            print("✅ Notification sent to Discord.")
        else:
            print(f"⚠️ Failed to send notification: {response.status_code}")
    except Exception as e:
        print(f"❌ Discord error: {e}")

# Fungsi ekstraksi file
def extract_file(file_path, extract_to):
    try:
        if file_path.endswith('.zip'):
            with zipfile.ZipFile(file_path, 'r') as zip_ref:
                zip_ref.extractall(extract_to)
            return True, 'ZIP extracted'

        elif file_path.endswith('.rar'):
            with rarfile.RarFile(file_path, 'r') as rar_ref:
                rar_ref.extractall(extract_to)
            return True, 'RAR extracted'

        elif file_path.endswith('.7z'):
            with py7zr.SevenZipFile(file_path, 'r') as seven_ref:
                seven_ref.extractall(extract_to)
            return True, '7Z extracted'

        else:
            return False, 'Not an archive file'

    except Exception as e:
        return False, f'Extraction failed: {e}'

# Handler untuk file baru
@client.on(events.NewMessage(chats=[-1813631420, -2105943934, -4631171373, -1508025016]))
async def handler(event):
    message = event.message
    if not message.file or not message.file.name:
        return

    filename = message.file.name.lower()
    valid_ext = ('.txt', '.zip', '.rar', '.7z')

    if filename.endswith(valid_ext):
        file_path = os.path.join(download_folder, message.file.name)
        print(f"⬇️ Downloading: {message.file.name}")
        await message.download_media(file_path)
        print(f"✅ File downloaded to: {file_path}")

        send_discord_notification(f"📥 Downloaded: `{message.file.name}`")

        # Ekstraksi otomatis jika archive
        if filename.endswith(('.zip', '.rar', '.7z')):
            success, info = extract_file(file_path, download_folder)
            if success:
                print(f"✅ {info} successfully.")
                send_discord_notification(f"🗂️ Extracted: `{message.file.name}` → {info}")
            else:
                print(f"⚠️ {info}")
                send_discord_notification(f"⚠️ Failed to extract `{message.file.name}`: {info}")

# Main loop
async def main():
    await client.start(phone_number)
    print("🚀 Bot is running and monitoring channels...")
    while True:
        await asyncio.sleep(86400)

with client:
    client.loop.run_until_complete(main())
