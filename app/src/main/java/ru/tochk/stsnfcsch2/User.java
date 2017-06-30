package ru.tochk.stsnfcsch2;

import android.util.JsonReader;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.net.MalformedURLException;
import java.net.URL;
import java.net.URLConnection;

/**
 * Created by tochk on 6/18/2017.
 */

class User {
    public String fullName;
    public String position;
    public String isStart;

    void getUserInfo(String cardId) throws IOException {
        URL server = new URL("http", "nfc.ssu.tochk.ru", 80, "submit/"+cardId);
        InputStream in = server.openStream();
        BufferedReader reader = new BufferedReader( new InputStreamReader( in )  );

        this.fullName = "Карта не найдена";
        this.isStart = "notfound";
        this.position = "Пожалуйста, обратитесь к администратору для регистрации карты";

        JsonReader rdr = new JsonReader(reader);
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
