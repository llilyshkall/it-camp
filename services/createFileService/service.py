from fastapi import FastAPI, Body, BackgroundTasks
from fastapi.responses import JSONResponse
from fastapi.middleware.cors import CORSMiddleware
import uvicorn
import os


app = FastAPI()

async def main():
    # Настройка CORS
    app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
    )

# Эндпоинт для проверки работы сервера
@app.get("/health")
async def health_check():
    return {"status": "ok"}



@app.post("/remarks_report")
async def remarksHandler(
    data: dict = Body(...)):
    #file = await checklist(projects_to_process)
    # TODO запись в S3
    return JSONResponse(content=None, status_code=200)

if __name__ == "__main__":
    #projects_to_process = ["Project_Alfa"]
    uvicorn.run(app, host="127.0.0.1", port=8086)