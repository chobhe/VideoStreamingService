import sys
import os
import cv2
import numpy as np
sys.path.append("..")

import requests
import json

import openai
from dotenv import load_dotenv

from fastapi import FastAPI, File, UploadFile, HTTPException
from fastapi.middleware.cors import CORSMiddleware
# used to return a json in the response
from fastapi.responses import JSONResponse
from fastapi.encoders import jsonable_encoder

#used to receive a json
from pydantic import BaseModel

from functions import parse_str, prompt_gen, closest_word, JsonReturn, split_module


'''
general setup
'''
load_dotenv()
server_endpt = os.environ['SERVER_IP']

'''
Init server
'''
app = FastAPI()

origins = [
    '*'
]

app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

    



@app.post('/label-faces')
async def label_faces(file: UploadFile = File(...)) -> JsonReturn :
    #source ./env/bin/activate to start venv
    # Ensure the file is in the right format
    if not file.filename.endswith(".flv"):
        raise HTTPException(status_code=400, detail="Invalid file format")

    # Save the file to disk temporarily, you may also process it directly in memory for larger files
    with open(f"/path/to/save/{file.filename}", "wb") as buffer:
        buffer.write(await file.read())

    # Use OpenCV to process the video file
    frames = []
    cap = cv2.VideoCapture(f"/{file.filename}")
    while cap.isOpened():
        ret, frame = cap.read()
        if not ret:
            break

        # frame is a numpy array containing an image
        frames.append(frame)

    cap.release()

    # Now you can process the frames with your computer vision code
    # ...

    return {"message": "Video processed successfully"}

   

    


# TODO
# def draw_boxes():

# def recompress data():