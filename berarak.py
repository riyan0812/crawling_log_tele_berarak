from telethon.sync import TelegramClient
from telethon import events
import os
import requests
import asyncio

# API dan nomor telepon
api_id = 'xxxxxxx'  # Ganti dengan API ID Anda
api_hash = 'xxxxxxx'  # Ganti dengan API Hash Anda
phone_number = 'xxxxxxxxx'  # Nomor telepon Anda yang terhubung ke Telegram

# Folder tempat file akan disimpan
download_folder = '/xxxxxx'

# Pastikan folder /kumpulan ada
if not os.path.exists(download_folder):
    os.makedirs(download_folder)

client = TelegramClient('session_name', api_id, api_hash)

# URL webhook Discord Anda
discord_webhook_url = 'xxxxxxxxxxxxxxx'  # Ganti dengan URL webhook Discord Anda

# Fungsi untuk mengirim notifikasi ke Discord
def send_discord_notification(message):
    data = {
        "content": message  # Pesan yang ingin dikirim ke Discord
    }
    response = requests.post(discord_webhook_url, json=data)
    if response.status_code == 204:
        print("Notification sent to Discord!")
    else:
        print(f"Failed to send notification: {response.status_code}")

# Fungsi untuk menangani file yang diupload
@client.on(events.NewMessage(chats=[xxxxxxxxxxxxxx]))  # Ganti dengan ID grup yang sesuai
async def handler(event):
    message = event.message
    # Mengecek apakah pesan berisi media dan memiliki ekstensi .txt
    if message.media and message.file.name.endswith('.txt'):
        print(f"Downloading: {message.file.name}")
        # Menentukan path lengkap untuk file yang akan disimpan
        file_path = os.path.join(download_folder, message.file.name)
        # Menyimpan file ke folder yang ditentukan
        await message.download_media(file_path)
        print(f"File downloaded to: {file_path}")

        # Kirim notifikasi ke Discord setelah file selesai diunduh
        send_discord_notification(f"File {message.file.name} has been downloaded successfully!")

# Fungsi utama
async def main():
    await client.start(phone_number)
    print("Bot is now running...")

    # Menjaga client berjalan selama 24 jam
    while True:
        await asyncio.sleep(86400)  # Menunggu selama 24 jam (86400 detik)

with client:
    client.loop.run_until_complete(main())
