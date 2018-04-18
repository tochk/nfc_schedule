package ru.tochk.nfc_schedule;

import android.annotation.SuppressLint;
import android.content.Intent;
import android.nfc.NfcAdapter;
import android.nfc.Tag;
import android.os.AsyncTask;
import android.os.Bundle;
import android.support.v7.app.AppCompatActivity;
import android.util.JsonReader;
import android.util.Log;
import android.view.View;
import android.widget.LinearLayout;
import android.widget.ProgressBar;
import android.widget.TextView;
import android.widget.Toast;

import com.blogspot.android_er.androidnfctagdiscovered.R;

import java.io.BufferedReader;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.net.URL;
import java.util.Arrays;

public class MainActivity extends AppCompatActivity {
    public TextView textViewInfo;
    private NfcAdapter nfcAdapter;
    @SuppressLint("SetTextI18n")
    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);
        textViewInfo = findViewById(R.id.info);
        nfcAdapter = NfcAdapter.getDefaultAdapter(this);
        if (nfcAdapter == null) {
            Toast.makeText(this,
                    "NFC NOT supported on this device!",
                    Toast.LENGTH_LONG).show();
            finish();
        } else if (!nfcAdapter.isEnabled()) {
            Toast.makeText(this,
                    "NFC NOT Enabled!",
                    Toast.LENGTH_LONG).show();
            finish();
        }
        Bundle b = getIntent().getExtras();
        if (b != null) {
            boolean ok = b.getBoolean("ok");
            if (ok) {
                TextView userViewInfo = findViewById(R.id.user);
                TextView positionViewInfo = findViewById(R.id.position);
                userViewInfo.setText(b.getString("fullName"));
                positionViewInfo.setText(b.getString("position"));
                textViewInfo.setText("Идентификатор карты\n" + b.getString("tag") + "\n\n");
                switch (b.getString("isStart")) {
                    case "true":
                        positionViewInfo.getRootView().setBackgroundColor(getResources().getColor(android.R.color.holo_green_light));
                        break;
                    case "false":
                        positionViewInfo.getRootView().setBackgroundColor(getResources().getColor(android.R.color.holo_orange_light));
                        break;
                    case "notfound":
                        positionViewInfo.getRootView().setBackgroundColor(getResources().getColor(android.R.color.holo_red_light));
                        break;
                }
            }
        } else {
            ProgressBar progressBar = findViewById(R.id.progress_bar);
            LinearLayout main = findViewById(R.id.main_ll);
            progressBar.setVisibility(View.VISIBLE);
            main.setVisibility(View.GONE);
        }

    }

    @SuppressLint("SetTextI18n")
    @Override
    protected void onResume() {
        super.onResume();

        Intent intent = getIntent();
        String action = intent.getAction();

        if (NfcAdapter.ACTION_TAG_DISCOVERED.equals(action)) {
            new CheckCard().execute();
            Tag tag = intent.getParcelableExtra(NfcAdapter.EXTRA_TAG);
            byte[] tagId = tag.getId();
            String tempTag;
            StringBuilder tempTagFull = new StringBuilder();
            for (byte aTagId : tagId) {
                tempTag = Integer.toHexString(aTagId & 0xFF);
                if (tempTag.length() == 1) {
                    tempTagFull.append("0").append(tempTag);
                } else if (tempTag.length() == 0) {
                    tempTagFull.append("00");
                } else {
                    tempTagFull.append(tempTag);
                }
            }
            //textViewInfo.setText("Идентификатор карты\n" + tempTagFull.toString() + "\n\n");
            new CheckCard().execute(tempTagFull.toString());
        }
    }

    @Override
    protected void onPause() {
        super.onPause();
        nfcAdapter = NfcAdapter.getDefaultAdapter(this);
    }


    @SuppressLint("StaticFieldLeak")
    class CheckCard extends AsyncTask<String, Void, String> {
        ProgressBar progressBar = findViewById(R.id.progress_bar);
        LinearLayout main = findViewById(R.id.main_ll);
        String fullName = "Карта не найдена";
        String position = "notfound";
        String isStart = "Пожалуйста, обратитесь к администратору для регистрации карты";
        String tag;

        protected void onPreExecute() {
            progressBar.setVisibility(View.VISIBLE);
            main.setVisibility(View.GONE);
        }

        protected String doInBackground(String... tagId) {
            try {
                this.tag = tagId[0];
                URL server = new URL("http", "nfc.ssu.tochk.ru", 80, "submit/" + this.tag);
                InputStream in = server.openStream();
                BufferedReader reader = new BufferedReader(new InputStreamReader(in));
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
            } catch (Exception e) {
                return null;
            }
            return null;
        }

        @SuppressLint("SetTextI18n")
        protected void onPostExecute(String response) {
            Intent intent = new Intent(MainActivity.this, MainActivity.class);
            Bundle b = new Bundle();
            b.putBoolean("ok", true);
            b.putString("fullName", this.fullName);
            b.putString("position", this.position);
            b.putString("isStart", this.isStart);
            b.putString("tag", this.tag);
            intent.putExtras(b);
            startActivity(intent);
        }
    }

}