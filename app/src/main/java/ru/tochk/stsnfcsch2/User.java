package ru.tochk.stsnfcsch2;

import android.util.JsonReader;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.net.MalformedURLException;
import java.net.URL;

/**
 * Created by tochk on 6/18/2017.
 */

class User {
    public String fullName;
    public String position;
    public String isStart;

    void getUserInfo(String cardId) throws IOException {
        URL oracle = new URL("http://neptun.ptr.sgu.ru:4005/" + cardId);
        BufferedReader in = new BufferedReader(new InputStreamReader(oracle.openStream()));

        JsonReader rdr = new JsonReader(in);
        rdr.beginObject();
        while (rdr.hasNext()) {
            String name = rdr.nextName();
            switch (name) {
                case "full_name":
                    this.fullName = rdr.nextString();
                    break;
                case "position":
                    this.position = rdr.nextString();
                    break;
                case "is_start":
                    this.isStart = rdr.nextString();
                    break;
                default:
                    rdr.skipValue();
                    break;
            }
        }
        rdr.endObject();
        in.close();
    }
}
