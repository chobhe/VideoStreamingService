import sys
import os
import cv2
import numpy as np
import io
sys.path.append("..")

import requests
import json

import openai
from dotenv import load_dotenv

from fastapi import FastAPI, HTTPException, Request
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
async def label_faces(request: Request) -> JsonReturn :
    #source ./env/bin/activate to start venv
    # Create a bytes buffer for the video
    # TODO: Perhaps create a more elegant streaming solution for receiving all the bytes instead of just
    # Reading it into memory: EG Process in chunks based on the frames or something
    body_bytes = await request.body()
    video_buffer = io.BytesIO(body_bytes)

    # Set up a list to hold the frames
    frames = []

    # Read the video buffer
    video_bytes = np.frombuffer(video_buffer.getvalue(), dtype=np.uint8)

    # Check for empty buffer
    if video_bytes is None:
        raise HTTPException(status_code=400, detail="Empty buffer")

    # Use OpenCV to process the video buffer
    cap = cv2.imdecode(video_bytes, cv2.IMREAD_COLOR)
    while cap.isOpened():
        ret, frame = cap.read()
        if not ret:
            break

        # frame is a numpy array containing an image
        frames.append(frame)

    cap.release()

    # Now you can process the frames with your computer vision code
    # Feed all the frames into a CV model

    return {"message": "Video processed successfully"}
    

   

    


# TODO
# def draw_boxes():

# def recompress data():