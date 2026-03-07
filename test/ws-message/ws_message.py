import asyncio
from asyncio.tasks import sleep
import websockets
import message_pb2
from datetime import datetime

async def test():
    uri = "ws://localhost:8001/ws"
    async with websockets.connect(uri, close_timeout=60, ping_interval=60, additional_headers={"PeerID":"203b7596-40ef-4e29-82af-17a790cdf68c"}) as ws:
        # Создаем сообщение
        msg = message_pb2.Message()
        msg.send_time = datetime.now()
        msg.message = "Test message"

        # Сериализуем в байты
        data = msg.SerializeToString()

        # Отправляем (binary=True обычно по умолчанию для bytes)
        for _ in range(10):
            await ws.send(data)
            await sleep(1)

        # Ждем ответ
        # response = await ws.recv()
        # print(f"Received {len(response)} bytes")

asyncio.run(test())
